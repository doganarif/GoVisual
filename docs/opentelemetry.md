# Using GoVisual with OpenTelemetry

OpenTelemetry export lives in its own module, `github.com/doganarif/govisual/telemetry`, so the OTel SDK and its gRPC stack stay out of builds that don't use them.

## Prerequisites

1. An OTLP-compatible collector running (such as Jaeger with OTLP enabled)
2. The telemetry module: `go get github.com/doganarif/govisual/telemetry`

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

### 2. Wrap your handler

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/doganarif/govisual/telemetry"
    "github.com/doganarif/govisual/v2"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", userHandler)

    // Trace the application handler. Wrapping the mux (not the final
    // handler) keeps the govisual dashboard out of your traces.
    traced, shutdown, err := telemetry.Wrap(mux, telemetry.Config{
        ServiceName:    "my-service",
        ServiceVersion: "1.0.0",
        Endpoint:       "localhost:4317",
        Insecure:       true,
        Exporter:       telemetry.ExporterOTLP,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(context.Background())

    handler := govisual.Wrap(traced,
        govisual.WithRequestBodyLogging(true),
        govisual.WithResponseBodyLogging(true),
    )

    http.ListenAndServe(":8080", handler)
}
```

### 3. Run the example

```bash
cd cmd/examples/otel
docker-compose up -d   # starts Jaeger
go run main.go
```

Traces appear in the Jaeger UI at http://localhost:16686.

## Configuration

`telemetry.Config` fields:

| Field            | Description                                         | Default |
| ---------------- | --------------------------------------------------- | ------- |
| `ServiceName`    | Service name attached to every span                 |         |
| `ServiceVersion` | Service version attached to every span              |         |
| `Endpoint`       | OTLP gRPC endpoint                                  |         |
| `Insecure`       | Skip TLS for the OTLP connection (local dev)        | `false` |
| `Exporter`       | `ExporterOTLP`, `ExporterStdout`, or `ExporterNoop` |         |

`telemetry.Wrap` returns the instrumented handler and a shutdown function that flushes pending spans — call it on exit. For finer control, `telemetry.InitTracer(ctx, cfg)` sets up the exporter and `telemetry.NewMiddleware(handler, name, version)` instruments a handler separately.

## Span propagation

The middleware extracts incoming W3C trace context headers, so spans join distributed traces started upstream. Handlers can start child spans with the global tracer:

```go
ctx, span := otel.Tracer("my-service").Start(r.Context(), "database.query")
defer span.End()
```

## Related Documentation

- [Configuration Options](configuration.md)
- [Quick Start Guide](quick-start.md)
