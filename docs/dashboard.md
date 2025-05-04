# GoVisual Dashboard

The GoVisual dashboard provides a real-time view of HTTP requests flowing through your application.

![GoVisual Dashboard](dashboard.png)

## Accessing the Dashboard

By default, the dashboard is available at `http://localhost:<your-port>/__viz`. You can customize this path using the `WithDashboardPath` option:

```go
handler := govisual.Wrap(
    mux,
    govisual.WithDashboardPath("/__debug"),
)
```

## Dashboard Features

### Request Table

The main view displays a table of recent HTTP requests with the following information:

- **Method**: HTTP method (GET, POST, PUT, etc.)
- **Path**: The request URL path
- **Status**: HTTP status code (color-coded)
- **Time**: Response time in milliseconds
- **Timestamp**: When the request was received

The table automatically updates as new requests come in.

### Request Details

Clicking on a request in the table reveals detailed information:

#### Request Tab

- Full URL (including query parameters)
- HTTP method
- Headers
- Request body (if enabled)
- Cookies

#### Response Tab

- Status code
- Headers
- Response body (if enabled)
- Content type
- Response size

#### Timing Tab

- Total response time
- Time spent in each middleware
- Network latency

#### Middleware Trace Tab

- Visual representation of middleware execution
- Time spent in each middleware
- Call hierarchy

### Filtering and Searching

The dashboard includes filtering capabilities:

- Filter by HTTP method (GET, POST, etc.)
- Filter by status code or status code range (2xx, 4xx, etc.)
- Search by URL path
- Filter by time range

### Dashboard Controls

- **Clear All**: Remove all requests from the view
- **Auto-refresh**: Toggle automatic updates
- **Columns**: Show/hide specific columns
- **Export**: Download request data as JSON

## Browser Support

The GoVisual dashboard is compatible with all modern browsers:

- Chrome (recommended)
- Firefox
- Safari
- Edge

## Troubleshooting

If you can't access the dashboard:

1. Verify that the dashboard path is correct
2. Check that your application is running
3. Ensure no path conflict with your application's routes
4. Check if any security middleware is blocking access

If requests aren't showing up:

1. Ensure the routes are passing through the GoVisual middleware
2. Check if the routes are in the ignored paths list
3. Make sure you're sending requests to the instrumented handler

## Related Documentation

- [Configuration Options](configuration.md) - Configure dashboard behavior
- [Request Logging](request-logging.md) - Control what gets logged
- [Middleware Tracing](middleware-tracing.md) - How middleware tracing works
