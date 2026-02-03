package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// ExporterOTLP is the standard OTLP gRPC exporter
	ExporterOTLP = "otlp"
	// ExporterStdout is the stdout exporter for debugging
	ExporterStdout = "stdout"
	// ExporterNoop is a no-operation exporter for benchmarking
	ExporterNoop = "noop"
)

// noopExporter is a SpanExporter that does nothing.
// Useful for benchmarking tracing overhead without network I/O.
type noopExporter struct{}

func (e *noopExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *noopExporter) Shutdown(_ context.Context) error {
	return nil
}

// Config holds the configuration for the telemetry package
type Config struct {
	ServiceName    string
	ServiceVersion string
	Endpoint       string
	Insecure       bool
	Exporter       string
}

// InitTracer initializes an OTLP exporter, and configures the corresponding trace provider.
func InitTracer(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	var traceExporter sdktrace.SpanExporter

	switch cfg.Exporter {
	case ExporterStdout:
		traceExporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}
	case ExporterNoop:
		traceExporter = &noopExporter{}
	case ExporterOTLP, "":
		// If no endpoint is provided, use a sensible default for local development
		endpoint := cfg.Endpoint
		if endpoint == "" {
			endpoint = "localhost:4317"
		}

		var opts []grpc.DialOption
		if cfg.Insecure {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
		}
		opts = append(opts, grpc.WithBlock())

		// Create gRPC connection to collector
		conn, err := grpc.DialContext(ctx, endpoint, opts...)
		if err != nil {
			return nil, err
		}

		// Create OTLP exporter
		traceExporter, err = otlptrace.New(ctx,
			otlptracegrpc.NewClient(
				otlptracegrpc.WithGRPCConn(conn),
			),
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown exporter type: %q (valid: %s, %s, %s)", cfg.Exporter, ExporterOTLP, ExporterStdout, ExporterNoop)
	}

	// Create trace provider with the exporter
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (W3C)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a shutdown function that can be called to clean up resources
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
			return err
		}
		return nil
	}, nil
}
