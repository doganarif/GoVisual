# Middleware Tracing Example

This example demonstrates the comprehensive middleware tracking capabilities of GoVisual.

## Features Demonstrated

- **Middleware Chain Tracking**: See how requests flow through multiple middleware layers
- **SQL Query Tracking**: Monitor database queries with timing and results
- **HTTP Call Tracking**: Track external API calls
- **Custom Trace Points**: Add custom trace entries for specific operations
- **Performance Metrics**: View detailed performance data for each request

## Running the Example

```bash
go run main.go
```

The server will start on http://localhost:8090

## Dashboard

Open http://localhost:8090/\_\_viz to view the dashboard.

## Test Endpoints

1. **Basic Request**: http://localhost:8090/

   - Simple JSON response
   - Shows basic middleware execution

2. **Database Query**: http://localhost:8090/api/users

   - Executes SQL queries
   - Shows database interaction in traces

3. **Slow Operation**: http://localhost:8090/api/slow

   - Multi-step operation with timing
   - Shows nested trace entries

4. **External API**: http://localhost:8090/api/external
   - Simulates external HTTP calls
   - Shows HTTP tracking in traces

## Viewing Traces

1. Make some requests to the test endpoints
2. Open the dashboard at http://localhost:8090/\_\_viz
3. Go to the "Trace" tab
4. Select a request to see its detailed execution trace

## Trace Information Includes

- **Middleware Stack**: See each middleware that processed the request
- **Execution Timeline**: Visual timeline of all operations
- **SQL Queries**: Full query text, duration, and row counts
- **HTTP Calls**: External API calls with status and timing
- **Custom Events**: Application-specific trace points
- **Performance Metrics**: CPU usage, memory allocation, and bottlenecks

## Customizing Traces

You can add custom trace points in your handlers:

```go
tracer := middleware.GetTracer(r.Context())
if tracer != nil {
    tracer.StartTrace("Operation Name", "custom", map[string]interface{}{
        "custom_field": "value",
    })
    // ... your operation ...
    tracer.EndTrace(nil)
}
```

## SQL Query Tracking

SQL queries are automatically tracked when using the profiling-enabled database drivers. The traces show:

- Query text
- Execution duration
- Number of rows affected/returned
- Any errors encountered
