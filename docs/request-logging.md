# Request Logging

GoVisual can capture and log HTTP requests and responses passing through your application. This document explains how request logging works and how to configure it.

## Basic Logging

By default, GoVisual logs basic request and response metadata:

- HTTP method (GET, POST, PUT, etc.)
- URL path
- Query parameters
- Status code
- Response time
- Timestamp
- Request and response headers

This basic information is always captured and does not require any special configuration.

## Body Logging

For more detailed logging, you can enable request and response body logging:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithRequestBodyLogging(true),  // Log request bodies
    govisual.WithResponseBodyLogging(true), // Log response bodies
)
```

### Important Considerations

When enabling body logging, keep in mind:

1. **Performance Impact**: Logging bodies requires reading them completely into memory, which may impact performance for large payloads
2. **Security Concerns**: Request and response bodies may contain sensitive information (passwords, tokens, PII)
3. **Memory Usage**: Bodies are stored in memory by default, which can increase memory usage

## Ignoring Paths

To prevent logging of certain paths (like health checks or static assets), use the `WithIgnorePaths` option:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithIgnorePaths(
        "/health",      // Exact match
        "/metrics",     // Exact match
        "/static/*",    // Wildcard pattern
        "/api/auth/*"   // Wildcard pattern
    ),
)
```

The dashboard path (`/__viz` by default) is automatically ignored to prevent recursive logging.

## Storage Considerations

How requests are stored depends on your configured storage backend:

- **Memory Storage**: Logs are kept in memory and lost when the application restarts
- **PostgreSQL Storage**: Logs are stored in a database table and persist across restarts
- **Redis Storage**: Logs are stored with a configurable time-to-live (TTL)

See [Storage Backends](storage-backends.md) for more details.

## Custom Headers

All headers are logged by default. If some headers contain sensitive information, you should handle them at the application level before they reach GoVisual.

## Request Log Format

Internally, GoVisual stores request logs with the following structure:

```go
type RequestLog struct {
    ID             string                 // Unique identifier
    Timestamp      time.Time              // When the request was received
    Method         string                 // HTTP method
    Path           string                 // URL path
    Query          string                 // Query parameters
    RequestHeaders map[string][]string    // Request headers
    ResponseHeaders map[string][]string   // Response headers
    StatusCode     int                    // HTTP status code
    Duration       time.Duration          // Response time
    RequestBody    string                 // Request body (if enabled)
    ResponseBody   string                 // Response body (if enabled)
    Error          string                 // Error message (if any)
    MiddlewareTrace []MiddlewareTraceEntry // Middleware execution trace
    RouteTrace     []RouteTraceEntry      // Route matching trace
}
```

## Example

A complete example of request logging configuration:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithRequestBodyLogging(true),
    govisual.WithResponseBodyLogging(true),
    govisual.WithIgnorePaths("/health", "/metrics", "/static/*"),
    govisual.WithMaxRequests(1000),
)
```

## Related Documentation

- [Configuration Options](configuration.md) - All available configuration options
- [Storage Backends](storage-backends.md) - Configure where logs are stored
- [Middleware Tracing](middleware-tracing.md) - How middleware tracing works
