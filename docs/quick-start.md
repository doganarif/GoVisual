# Quick Start Guide

This guide will help you integrate GoVisual into your Go web application in just a few minutes.

## Basic Integration

The core concept of GoVisual is simple - wrap your existing HTTP handler with `govisual.Wrap()`:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/doganarif/govisual"
)

func main() {
    // Create your regular HTTP handler/mux
    mux := http.NewServeMux()

    // Add your routes
    mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })

    // Wrap with GoVisual (with default settings)
    handler := govisual.Wrap(mux)

    // Start your server with the wrapped handler
    http.ListenAndServe(":8080", handler)
}
```

That's it! Your application will now have a built-in request visualization dashboard accessible at `http://localhost:8080/__viz`.

## Configuration Options

For more control, you can provide options to the `Wrap` function:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(100),             // Store the last 100 requests
    govisual.WithDashboardPath("/__debug"),    // Custom dashboard path
    govisual.WithRequestBodyLogging(true),     // Log request bodies
    govisual.WithResponseBodyLogging(true),    // Log response bodies
    govisual.WithIgnorePaths("/health", "/metrics") // Don't log these paths
)
```

## Accessing the Dashboard

By default, the dashboard is available at `http://localhost:8080/__viz` (or whatever port your application is running on).

The dashboard shows:

- A list of all captured HTTP requests
- Detailed information about each request and response
- Request and response bodies (if enabled)
- Middleware execution trace
- Timing information

## Next Steps

- [Configuration Options](configuration.md) - Learn about all available configuration options
- [Storage Backends](storage-backends.md) - Configure persistent storage for request logs
- [OpenTelemetry Integration](opentelemetry.md) - Export trace data to OpenTelemetry collectors
- [Examples](../cmd/examples/basic/README.md) - Run the included examples
