# Using GoVisual with OpenTelemetry

GoVisual includes OpenTelemetry integration, allowing you to export telemetry data to your preferred backend.

## Prerequisites

To use OpenTelemetry with GoVisual, you need:

1. An OTLP-compatible collector running (such as Jaeger with OTLP enabled)
2. GoVisual v1.0.0 or later

## Quick Start

### 1. Start an OpenTelemetry Backend

The easiest way to get started is with Jaeger. Run it using Docker:

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### 2. Enable OpenTelemetry in your Go application

```go
package main

import (
    "net/http"
    "github.com/doganarif/govisual"
)

func main() {
    mux := http.NewServeMux()

    // Add your routes
    mux.HandleFunc("/api/users", userHandler)

    // Enable GoVisual with OpenTelemetry
    handler := govisual.Wrap(
        mux,
        govisual.WithOpenTelemetry(true),               // Enable OpenTelemetry
        govisual.WithServiceName("my-service"),         // Set service name
        govisual.WithServiceVersion("1.0.0"),           // Set service version
        govisual.WithOTelEndpoint("localhost:4317"),    // OTLP exporter endpoint
        govisual.WithRequestBodyLogging(true),          // Log request bodies
        govisual.WithResponseBodyLogging(true),         // Log response bodies
    )

    http.ListenAndServe(":8080", handler)
}
```

### 3. Run the example

You can run the included example in the repository:

```bash
cd cmd/examples/otel
docker-compose up -d
go run main.go
```

Then visit:

- GoVisual dashboard: http://localhost:8080/\_\_viz
- Jaeger UI: http://localhost:16686

## Configuration Options

| Option                       | Description                            | Default            |
| ---------------------------- | -------------------------------------- | ------------------ |
| `WithOpenTelemetry(bool)`    | Enable OpenTelemetry integration       | `false`            |
| `WithServiceName(string)`    | Set the service name for OpenTelemetry | `"govisual"`       |
| `WithServiceVersion(string)` | Set the service version                | `"dev"`            |
| `WithOTelEndpoint(string)`   | Set the OTLP exporter endpoint         | `"localhost:4317"` |

## How It Works

When OpenTelemetry is enabled:

1. GoVisual initializes an OpenTelemetry tracer with the provided service name and version
2. HTTP requests passing through GoVisual are automatically traced
3. Trace data is exported to the configured OTLP endpoint
4. The original GoVisual dashboard continues to work alongside OpenTelemetry

## Adding Custom Spans

You can add custom spans within your request handlers:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Get the context from the request (it contains the parent span from GoVisual)
    ctx := r.Context()

    // Start a new child span
    ctx, span := otel.Tracer("my-service").Start(ctx, "my-operation")
    defer span.End()

    // Add attributes to the span
    span.SetAttributes(attribute.String("user.id", "123"))

    // Create nested spans for detailed operations
    _, dbSpan := otel.Tracer("my-service").Start(ctx, "database.query")
    // ... do database work
    dbSpan.End()

    // Respond to the client
    w.Write([]byte("Hello, world!"))
}
```

## Troubleshooting

If you encounter issues:

1. Check that your OpenTelemetry collector is running and accessible
2. Ensure the endpoint in `WithOTelEndpoint()` matches your collector configuration
3. Look for initialization errors in your application logs

To see a full working example, refer to the `cmd/examples/otel` directory in the repository.
