# Frequently Asked Questions

## General Questions

### What is GoVisual?

GoVisual is a lightweight, zero-configuration HTTP request visualizer and debugger for Go web applications. It helps you monitor, inspect, and debug HTTP requests during local development.

### How does GoVisual work?

GoVisual works as middleware that wraps your existing HTTP handlers. It intercepts requests and responses, collects information about them, and provides a dashboard interface to visualize this data.

### Is GoVisual suitable for production use?

GoVisual is primarily designed for local development and testing environments. While it can be used in production for debugging purposes, we recommend enabling it selectively or using it with caution due to:

- Potential performance impact when logging all requests
- Memory usage when storing request and response bodies
- Security considerations for sensitive information

For production monitoring, consider using OpenTelemetry integration with a proper observability backend.

## Implementation

### Can I use GoVisual with any Go HTTP router/framework?

Yes, GoVisual works with any Go HTTP handler that implements the `http.Handler` interface, including:

- Standard library `http.ServeMux`
- Gorilla Mux
- Chi
- Echo (with adapter)
- Gin (with adapter)

### How do I use GoVisual with Echo/Gin/Fiber frameworks?

For frameworks that don't directly use the standard `http.Handler` interface, you'll need to use an adapter:

**Echo Example:**

```go
// Create Echo instance
e := echo.New()

// Add your routes
e.GET("/hello", helloHandler)

// Wrap Echo's HTTP handler with GoVisual
echoHandler := govisual.Wrap(echo.WrapHandler(e))

// Start server with the wrapped handler
http.ListenAndServe(":8080", echoHandler)
```

**Gin Example:**

```go
// Create Gin router
r := gin.Default()

// Add your routes
r.GET("/hello", helloHandler)

// Wrap Gin's HTTP handler with GoVisual
ginHandler := govisual.Wrap(r)

// Start server with the wrapped handler
http.ListenAndServe(":8080", ginHandler)
```

### Does GoVisual work with WebSockets?

GoVisual works with WebSocket handshake requests but does not trace the WebSocket communication itself after the connection is established.

## Configuration

### How do I change the dashboard URL?

Use the `WithDashboardPath` option:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithDashboardPath("/__debug"),
)
```

### Can I configure GoVisual from environment variables?

GoVisual doesn't directly read environment variables, but you can easily create your own configuration setup:

```go
func configureGoVisual(handler http.Handler) http.Handler {
    // Read config from environment
    maxRequests, _ := strconv.Atoi(getEnvOrDefault("GOVISUAL_MAX_REQUESTS", "100"))
    logBodies := getEnvOrDefault("GOVISUAL_LOG_BODIES", "false") == "true"
    dashPath := getEnvOrDefault("GOVISUAL_DASHBOARD_PATH", "/__viz")

    // Apply configuration
    return govisual.Wrap(
        handler,
        govisual.WithMaxRequests(maxRequests),
        govisual.WithRequestBodyLogging(logBodies),
        govisual.WithResponseBodyLogging(logBodies),
        govisual.WithDashboardPath(dashPath),
    )
}

func getEnvOrDefault(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
```

## Performance

### What is the performance impact of using GoVisual?

The performance impact depends on how GoVisual is configured:

- Basic request metadata logging: Minimal impact (typically < 1ms per request)
- Body logging: Impact scales with the size of request/response bodies
- Storage backend: In-memory is fastest, followed by Redis, then PostgreSQL
- Number of requests stored: Higher numbers require more memory

For best performance:

- Disable body logging (`WithRequestBodyLogging(false)`, `WithResponseBodyLogging(false)`)
- Use a smaller maximum request count (`WithMaxRequests(50)`)
- Use Redis for storage in high-volume applications

### How can I minimize memory usage?

To reduce memory usage:

- Disable body logging
- Reduce the number of stored requests
- Use Redis storage with a low TTL
- Ignore high-volume paths

## Storage

### Which storage backend should I choose?

- **In-memory**: Simplest, fastest, but data is lost on restart
- **Redis**: Good balance of performance and persistence, with automatic expiration
- **PostgreSQL**: Best for long-term storage and complex querying

### How do I implement a custom storage backend?

Implement the `Store` interface from `internal/store/store.go`:

```go
type Store interface {
    AddRequest(log *RequestLog) error
    GetRequest(id string) (*RequestLog, error)
    GetRequests() ([]*RequestLog, error)
    Clear() error
    Close() error
}
```

Then use it directly in your application.

## Other

### Does GoVisual work with HTTPS?

Yes, GoVisual works with both HTTP and HTTPS servers. It operates at the handler level, so the transport protocol doesn't matter.

### Can I use GoVisual with gRPC?

GoVisual is designed for HTTP traffic. While it doesn't directly support gRPC, you can use it alongside gRPC servers by running it on a different port.

### How is GoVisual different from other debugging tools?

Compared to other tools:

- **vs. net/http/pprof**: GoVisual focuses on HTTP request visualization, while pprof is for profiling CPU, memory, etc.
- **vs. debugcharts**: GoVisual traces HTTP requests, while debugcharts visualizes runtime statistics.
- **vs. APM tools**: GoVisual is lightweight and requires no external services, designed specifically for local development.

## Related Documentation

- [Configuration Options](configuration.md) - Available configuration options
- [Storage Backends](storage-backends.md) - Different storage options
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
