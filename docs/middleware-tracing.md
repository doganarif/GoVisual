# Middleware Tracing

GoVisual provides in-depth tracing of middleware execution in your HTTP request handling pipeline. This feature helps you visualize the flow of requests through your middleware stack and identify performance bottlenecks.

## How Middleware Tracing Works

When a request is processed through a GoVisual-wrapped handler, the middleware tracer:

1. Records when each middleware function starts and ends
2. Captures the time spent in each middleware component
3. Builds a hierarchical representation of the middleware stack
4. Visualizes this data in the dashboard

## Viewing Middleware Traces

In the GoVisual dashboard, click on any request to see its details, then navigate to the "Middleware Trace" tab. The trace is displayed as a hierarchical tree with:

- Middleware name/type
- Execution time (absolute and relative)
- Execution order
- Parent-child relationships

## Supported Middleware

GoVisual can trace any standard Go HTTP middleware that follows the common middleware chaining patterns:

### Standard HTTP Middleware

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        next.ServeHTTP(w, r)
        // Post-processing
    })
}
```

### Middleware Using Context

```go
func ContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Add something to context
        ctx := context.WithValue(r.Context(), "key", "value")
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Function Adapters

```go
func LoggingAdapter(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Log request
        h(w, r)
        // Log response
    }
}
```

## Middleware Trace Format

Internally, middleware traces are stored with this structure:

```go
type MiddlewareTraceEntry struct {
    Name      string            // Name of the middleware
    StartTime time.Time         // When middleware execution started
    EndTime   time.Time         // When middleware execution ended
    Duration  time.Duration     // Time spent in this middleware
    Depth     int               // Nesting level in the middleware stack
    Parent    int               // Index of parent middleware (-1 for root)
    Children  []int             // Indices of child middleware entries
    Metadata  map[string]string // Additional metadata
}
```

## Performance Considerations

Middleware tracing adds minimal overhead to request processing:

- Typically less than 0.1ms per middleware layer
- Trace data is collected only for successful requests
- Trace size scales with middleware complexity

## Example: Analyzing Middleware Performance

Here's a common scenario where middleware tracing helps:

1. You notice certain API requests are slower than expected
2. In the GoVisual dashboard, you find these requests in the table
3. You open the middleware trace and see one middleware taking significantly longer than others
4. You optimize that middleware and immediately see the performance improvement

## Relationship with OpenTelemetry

When OpenTelemetry integration is enabled:

- Middleware traces are exported as OpenTelemetry spans
- Each middleware becomes a span in the trace
- The hierarchy is preserved in the span relationship
- You can view the same data in your OpenTelemetry backend (Jaeger, Zipkin, etc.)

## Related Documentation

- [OpenTelemetry Integration](opentelemetry.md) - Exporting middleware traces to OpenTelemetry
- [Dashboard](dashboard.md) - How to use the dashboard to view middleware traces
- [Configuration Options](configuration.md) - Available configuration options
