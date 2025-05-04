# GoVisual

A lightweight, zero-configuration HTTP request visualizer and debugger for Go web applications during local development.

## Features

- **Real-time Request Monitoring**: Visualize HTTP requests passing through your application
- **Request Inspection**: Deep inspection of headers, body, status codes, and timing information
- **Middleware Tracing**: Visualize middleware execution flow and identify performance bottlenecks
- **Zero Configuration**: Drop-in integration with standard Go HTTP handlers
- **OpenTelemetry Integration**: Optional export of telemetry data to OpenTelemetry collectors

## Installation

```bash
go get github.com/doganarif/govisual
```

## Quick Start

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

    // Wrap with GoVisual
    handler := govisual.Wrap(
        mux,
        govisual.WithRequestBodyLogging(true),
        govisual.WithResponseBodyLogging(true),
    )

    http.ListenAndServe(":8080", handler)
}
```

Access the dashboard at `http://localhost:8080/__viz`

## Configuration Options

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(100),              // Number of requests to store
    govisual.WithDashboardPath("/__dashboard"), // Custom dashboard path
    govisual.WithRequestBodyLogging(true),      // Log request bodies
    govisual.WithResponseBodyLogging(true),     // Log response bodies
    govisual.WithIgnorePaths("/health"),        // Paths to ignore
    govisual.WithOpenTelemetry(true),           // Enable OpenTelemetry
    govisual.WithServiceName("my-service"),     // Service name for OTel
    govisual.WithServiceVersion("1.0.0"),       // Service version
    govisual.WithOTelEndpoint("localhost:4317"), // OTLP endpoint
)
```

## Examples

### Basic Example

Simple example showing core functionalities:

```bash
cd cmd/examples/basic
go run main.go
```

### OpenTelemetry Example

Example showing integration with OpenTelemetry:

```bash
cd cmd/examples/otel
docker-compose up -d  # Start Jaeger
go run main.go
```

Visit [OpenTelemetry Integration](docs/opentelemetry.md) for detailed instructions.

## Dashboard Features

![GoVisual Dashboard](docs/dashboard.png)

- **Request Table**: View all captured HTTP requests with method, path, status code, and response time
- **Request Details**: One-click access to headers, body content, and timing information
- **Middleware Trace**: Interactive visualization of middleware execution flow
- **Request Filtering**: Filter by HTTP method, status code, path pattern, or duration
- **Real-time Updates**: See new requests appear instantly as they happen

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
