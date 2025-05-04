# GoVisual

A lightweight, zero-configuration HTTP request visualizer and debugger for Go web applications during local development.

## Features

- **Developer-Friendly Dashboard**: View all HTTP requests in your local development environment
- **Real-time Request Monitoring**: Visualize HTTP requests passing through your application as they happen
- **Request Inspection**: Deep inspection of headers, body, status codes, and timing information
- **Middleware Tracing**: Visualize middleware execution flow and performance bottlenecks
- **Environment Info**: View Go runtime and system environment at a glance
- **Interactive UI**: Filter requests, explore details, and analyze request patterns
- **Zero Configuration**: Drop-in integration with standard Go HTTP handlers
- **Zero External Dependencies**: Fully self-contained with no third-party dependencies

## Installation

```bash
go get github.com/doganarif/govisual
```

## Quick Start

Just wrap your HTTP handler with GoVisual during development:

```go
package main

import (
    "net/http"
    "flag"
    
    "github.com/doganarif/govisual"
)

func main() {
    // Use a flag to enable the dashboard only in development
    var enableVisualizer bool
    flag.BoolVar(&enableVisualizer, "viz", false, "Enable request visualizer")
    flag.Parse()
    
    // Create your HTTP handler
    mux := http.NewServeMux()
    
    // Add your routes
    mux.HandleFunc("/api/users", userHandler)
    mux.HandleFunc("/api/products", productHandler)
    // ... your other routes
    
    var handler http.Handler = mux
    
    // Only wrap with GoVisual during development
    if enableVisualizer {
        handler = govisual.Wrap(
            mux,
            govisual.WithRequestBodyLogging(true),
            govisual.WithResponseBodyLogging(true),
        )
        println("🔍 Request visualizer enabled! Access at http://localhost:8080/__viz")
    }
    
    // Start the server
    http.ListenAndServe(":8080", handler)
}
```

Run with the visualizer during development:
```bash
go run main.go -viz
```

## Configuration Options

GoVisual can be configured with various options:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(100),              // Number of requests to store
    govisual.WithDashboardPath("/__dashboard"), // Custom dashboard path
    govisual.WithRequestBodyLogging(true),      // Log request bodies
    govisual.WithResponseBodyLogging(true),     // Log response bodies
    govisual.WithIgnorePaths("/health", "/metrics"), // Paths to ignore
)
```

## Dashboard Features

Once your application is running, access the dashboard at `http://localhost:8080/__viz`:

![GoVisual Dashboard](docs/dashboard.png)

- **Request Table**: View all captured HTTP requests with method, path, status code, and response time
- **Request Details**: One-click access to headers, body content (JSON formatted), and full timing information
- **Middleware Trace**: Interactive visualization of middleware execution flow showing bottlenecks
- **Request Filtering**: Quickly filter by HTTP method, status code, path pattern, or minimum duration
- **Environment Inspector**: Debug configuration with Go runtime and environment variable information
- **Real-time Updates**: See new requests appear instantly as they hit your application

## Architecture

GoVisual uses a modular architecture:

- **Middleware**: Captures HTTP requests and responses
- **Store**: Stores request logs in memory (with configurable capacity)
- **Dashboard**: Provides the visualization UI
- **Model**: Defines the data structures for request logs

## Development Workflow Benefits

- **Debug API Calls**: Easily inspect request and response bodies without external tools
- **Troubleshoot Middleware**: Identify slow middleware components in your processing chain
- **Test Different Clients**: Compare requests from different clients or API consumers
- **Monitor During Testing**: Keep the dashboard open while running integration tests
- **Share With Team**: Show request patterns to teammates during pair programming
- **Local Only**: Designed for development environments, not recommended for production

## Local Development Use Cases

### API Development
When building REST APIs or GraphQL endpoints, GoVisual helps you:
- Confirm request and response bodies match your API spec
- Debug client integration issues by inspecting exactly what was sent
- Compare request patterns between different API versions

### Middleware Debugging
Identify performance bottlenecks in your middleware chain:
- See which middleware components take the most time
- Track the execution order of complex middleware stacks
- Identify middleware that's modifying your request/response

### Framework Integration
GoVisual works with popular Go web frameworks:
- Standard library (`net/http`)
- Gin
- Echo
- Chi
- Fiber (with adapter)

Just wrap the router or mux provided by your framework.

## Example

See the `cmd/example` directory for a complete example implementation including middleware tracing.

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.