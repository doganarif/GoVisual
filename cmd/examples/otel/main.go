package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/doganarif/govisual/telemetry"
	"github.com/doganarif/govisual/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	var (
		port       int
		enableOTel bool
	)
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.BoolVar(&enableOTel, "otel", true, "Enable OpenTelemetry")

	var otelExporter string
	flag.StringVar(&otelExporter, "exporter", "otlp", "OpenTelemetry exporter (otlp, stdout, or noop)")

	flag.Parse()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Add routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/users", usersHandler)
	mux.HandleFunc("/api/search", searchHandler)
	mux.HandleFunc("/api/health", healthHandler)

	// Trace the application handler; wrapping the mux (not the final
	// handler) keeps the govisual dashboard out of the traces.
	var app http.Handler = mux
	if enableOTel {
		traced, shutdown, err := telemetry.Wrap(mux, telemetry.Config{
			ServiceName:    "govisual-otel-example",
			ServiceVersion: "1.0.0",
			Endpoint:       "localhost:4317",
			Insecure:       true,
			Exporter:       otelExporter,
		})
		if err != nil {
			log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
		}
		defer shutdown(context.Background())
		app = traced
		log.Printf("🔭 OpenTelemetry enabled with %s exporter!", otelExporter)
	}

	// Wrap with GoVisual
	handler := govisual.Wrap(app,
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithIgnorePaths("/api/health"),
	)
	log.Printf("🔍 Request visualizer enabled at http://localhost:%d/__viz", port)

	// Start the server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server started at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body>
		<h1>GoVisual OpenTelemetry Example</h1>
		<p>Visit <a href="/__viz">/__viz</a> to access the request visualizer</p>
		<p>Visit <a href="http://localhost:16686" target="_blank">Jaeger UI</a> to see traces</p>
		<p>API Endpoints:</p>
		<ul>
			<li><a href="/api/users">/api/users</a> - Get users with nested spans</li>
			<li><a href="/api/search?q=test">/api/search?q=test</a> - Search with attributes</li>
			<li><a href="/api/health">/api/health</a> - Health check (not traced)</li>
		</ul>
	</body></html>`)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := otel.Tracer("example").Start(ctx, "usersHandler",
		trace.WithAttributes(attribute.String("handler", "users")))
	defer span.End()

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	// Create child span without using the context
	_, childSpan := otel.Tracer("example").Start(ctx, "database.query")
	time.Sleep(150 * time.Millisecond)
	childSpan.End()

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{
		{"id": 1, "name": "John Doe"},
		{"id": 2, "name": "Jane Smith"},
	})
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("example").Start(r.Context(), "searchHandler")
	defer span.End()

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	// Add attribute to span
	span.SetAttributes(attribute.String("search.query", query))

	// Simulate search
	time.Sleep(200 * time.Millisecond)

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query": query,
		"results": []map[string]string{
			{"name": "Result 1"},
			{"name": "Result 2"},
		},
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}
