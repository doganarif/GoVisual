package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestOTelMiddleware_ServeHTTP(t *testing.T) {
	// Set up a test tracer provider and span recorder
	sr := sdktrace.NewSimpleSpanProcessor(&testSpanExporter{})
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	// Dummy handler to respond with status
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		io.WriteString(w, "traced")
	})

	// Wrap with OTelMiddleware
	middleware := NewOTelMiddleware(handler, "test-service", "v1.0")

	req := httptest.NewRequest("GET", "/hello", nil)
	req.Header.Set("User-Agent", "TestClient")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Basic output check
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

// testSpanExporter implements a basic SpanExporter to print spans (you can extend this to assert attributes)
type testSpanExporter struct{}

func (e *testSpanExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		if span.Name() == "" {
			panic("span name is empty")
		}
		if span.SpanKind() != trace.SpanKindServer {
			panic("span kind is not server")
		}
	}
	return nil
}

func (e *testSpanExporter) Shutdown(_ context.Context) error {
	return nil
}
