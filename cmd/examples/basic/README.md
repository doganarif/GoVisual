# Basic GoVisual Example

This example demonstrates the core functionality of GoVisual.

## Running the Example

```bash
go run main.go
```

This will start a server on port 8080 with the following endpoints:

- `/` - Home page with links to API endpoints
- `/api/hello` - Simple JSON response (100ms delay)
- `/api/slow` - Slow response (500ms delay)
- `/api/error` - Error response (500 status code)

## Accessing the Dashboard

Once the server is running, you can access the GoVisual dashboard at [http://localhost:8080/\_\_viz](http://localhost:8080/__viz).

## Features Demonstrated

- Request visualization
- Response timing
- Status code tracking
- Request and response body logging
