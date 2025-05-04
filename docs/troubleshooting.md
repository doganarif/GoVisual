# Troubleshooting

This guide covers common issues and their solutions when working with GoVisual.

## Dashboard Not Accessible

### Symptom

You can't access the GoVisual dashboard at the expected URL path.

### Possible Causes and Solutions

1. **Wrong Dashboard Path**

   - Confirm the dashboard path in your configuration
   - Default is `/__viz`, but you may have changed it with `WithDashboardPath`
   - Example: `http://localhost:8080/__viz`

2. **Application Not Running**

   - Verify your application is running
   - Check for any startup errors in logs

3. **Path Conflict**

   - Your application might have a route that conflicts with the dashboard path
   - Change the dashboard path to avoid conflicts:
     ```go
     handler := govisual.Wrap(mux, govisual.WithDashboardPath("/__govisual"))
     ```

4. **Middleware Not Applied**
   - Make sure the GoVisual middleware is correctly applied
   - The handler returned by `govisual.Wrap()` must be the one used by your server

## No Requests Showing in Dashboard

### Symptom

The dashboard is accessible, but no requests are displayed.

### Possible Causes and Solutions

1. **Ignored Paths**

   - Check if the paths you're accessing are in the ignored list
   - Default ignored path is the dashboard path itself

2. **Wrong Handler Chain**

   - Ensure that requests are passing through the GoVisual middleware
   - Common mistake: wrapping a handler that isn't used by your server

3. **Storage Issues**
   - If using a custom storage backend, check for connectivity issues
   - Verify that the storage is properly configured

## Performance Issues

### Symptom

Application becomes slow after integrating GoVisual.

### Possible Causes and Solutions

1. **Request Body Logging**

   - Disable request body logging for large payloads:
     ```go
     govisual.WithRequestBodyLogging(false)
     ```

2. **Response Body Logging**

   - Disable response body logging for large responses:
     ```go
     govisual.WithResponseBodyLogging(false)
     ```

3. **High Request Volume**

   - Reduce the number of stored requests:
     ```go
     govisual.WithMaxRequests(50) // Default is 100
     ```

4. **Memory Storage Limits**
   - Consider using Redis or PostgreSQL for high-volume applications

## Storage Backend Issues

### PostgreSQL Issues

1. **Connection Failed**

   - Check connection string
   - Verify the database exists and is accessible
   - Ensure user has appropriate permissions

2. **Table Not Created**
   - GoVisual should create the table automatically
   - If it fails, create the table manually:
     ```sql
     CREATE TABLE IF NOT EXISTS govisual_requests (
       id TEXT PRIMARY KEY,
       timestamp TIMESTAMP WITH TIME ZONE,
       method TEXT,
       path TEXT,
       query TEXT,
       request_headers JSONB,
       response_headers JSONB,
       status_code INTEGER,
       duration BIGINT,
       request_body TEXT,
       response_body TEXT,
       error TEXT,
       middleware_trace JSONB,
       route_trace JSONB,
       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
     )
     ```

### Redis Issues

1. **Connection Failed**

   - Check Redis connection string format
   - Verify Redis server is running
   - Test connection with Redis CLI

2. **Memory Issues**
   - Adjust TTL to clean up older entries faster
   - Example: `govisual.WithRedisStorage("redis://localhost:6379/0", 3600) // 1 hour TTL`

## OpenTelemetry Integration Issues

1. **Failed to Initialize OpenTelemetry**

   - Check logs for initialization errors
   - Verify the OTLP endpoint is correctly specified
   - Ensure the OpenTelemetry collector is running

2. **No Spans in Collector**
   - Verify collector configuration accepts OTLP format
   - For Jaeger, ensure the OTLP receiver is enabled
   - Try changing the endpoint format or port

## Middleware Tracing Issues

1. **Missing Middleware in Trace**

   - Some third-party middleware may not be properly detected
   - Ensure middleware follows standard Go patterns

2. **Incorrect Timing**
   - High system load can affect timing accuracy
   - Consider testing under normal load conditions

## Getting Help

If the above solutions don't resolve your issue:

1. Check the [GoVisual GitHub repository](https://github.com/doganarif/govisual) for open issues
2. Create a new issue with:
   - Detailed description of the problem
   - Error messages and logs
   - Configuration code
   - Steps to reproduce
   - Version information (Go, GoVisual, OS)

## Related Documentation

- [Configuration Options](configuration.md) - Review available configuration options
- [Storage Backends](storage-backends.md) - Storage backend configuration details
- [Frequently Asked Questions](faq.md) - Common questions and answers
