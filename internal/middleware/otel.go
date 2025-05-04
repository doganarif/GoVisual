package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// OTelMiddleware wraps an http.Handler with OpenTelemetry instrumentation
type OTelMiddleware struct {
	tracer         trace.Tracer
	propagator     propagation.TextMapPropagator
	handler        http.Handler
	serviceVersion string
}

// NewOTelMiddleware creates a new OpenTelemetry middleware
func NewOTelMiddleware(handler http.Handler, serviceName, serviceVersion string) *OTelMiddleware {
	return &OTelMiddleware{
		tracer:         otel.Tracer(serviceName),
		propagator:     otel.GetTextMapPropagator(),
		handler:        handler,
		serviceVersion: serviceVersion,
	}
}

// ServeHTTP implements the http.Handler interface
func (m *OTelMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract any existing context from the request
	ctx := r.Context()
	ctx = m.propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

	// Start a new span
	spanName := r.Method + " " + r.URL.Path
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.host", r.Host),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("http.flavor", r.Proto),
			attribute.String("service.version", m.serviceVersion),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	ctx, span := m.tracer.Start(ctx, spanName, opts...)
	defer span.End()

	// Create wrapped response writer to capture status code
	wrw := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Execute handler with context
	m.handler.ServeHTTP(wrw, r.WithContext(ctx))

	// Add status code attribute to span
	span.SetAttributes(attribute.Int("http.status_code", wrw.statusCode))
}

// wrappedResponseWriter captures the status code
type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (wrw *wrappedResponseWriter) WriteHeader(statusCode int) {
	wrw.statusCode = statusCode
	wrw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures writes to the response
func (wrw *wrappedResponseWriter) Write(b []byte) (int, error) {
	return wrw.ResponseWriter.Write(b)
}
