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

## Documentation

For detailed documentation, please refer to the [DOCS](docs/README.md).

## Configuration Options

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(100),              // Number of requests to store
    govisual.WithDashboardPath("/__dashboard"), // Custom dashboard path
    govisual.WithRequestBodyLogging(true),      // Log request bodies
    govisual.WithResponseBodyLogging(true),     // Log response bodies
    govisual.WithConsoleLogging(true),          // Log request timing to TTY
    govisual.WithIgnorePaths("/health"),        // Paths to ignore
    govisual.WithOpenTelemetry(true),           // Enable OpenTelemetry
    govisual.WithServiceName("my-service"),     // Service name for OTel
    govisual.WithServiceVersion("1.0.0"),       // Service version
    govisual.WithOTelEndpoint("localhost:4317"), // OTLP endpoint

    // Storage options (choose one)
    govisual.WithMemoryStorage(),                // In-memory storage (default)
    govisual.WithPostgresStorage(               // PostgreSQL storage
        "postgres://user:password@localhost:5432/database?sslmode=disable",
        "govisual_requests"                     // Table name
    ),
    govisual.WithRedisStorage(                  // Redis storage
        "redis://localhost:6379/0",             // Redis connection string
        86400                                   // TTL in seconds (24 hours)
    ),
)
```

## Storage Backends

GoVisual supports multiple storage backends for storing request logs:

### In-Memory Storage (Default)

The default storage keeps all request logs in memory. This is the simplest option but logs will be lost when the application restarts.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMemoryStorage(), // Optional, this is the default
)
```

### PostgreSQL Storage

For persistent storage, you can use PostgreSQL. This requires the `github.com/lib/pq` package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithPostgresStorage(
        "postgres://user:password@localhost:5432/dbname?sslmode=disable", // Connection string
        "govisual_requests"  // Table name (created automatically if it doesn't exist)
    ),
)
```

### Redis Storage

For high-performance storage with automatic expiration, you can use Redis. This requires the `github.com/go-redis/redis/v8` package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithRedisStorage(
        "redis://localhost:6379/0", // Redis connection string
        86400                       // TTL in seconds (24 hours)
    ),
)
```

## SQLite Driver Conflict

If you're already using a SQLite driver in your application (such as `github.com/mattn/go-sqlite3`), you may experience a conflict when using govisual with SQLite storage:

```
panic: sql: Register called twice for driver sqlite3
```

This occurs because both your application and govisual try to register the SQLite driver with the same name.

To resolve this issue, you can pass your existing database connection to govisual:

```go
import (
    "database/sql"

    "github.com/doganarif/govisual"
    _ "github.com/mattn/go-sqlite3" // Your preferred SQLite driver
)

func main() {
    // Create your own database connection
    db, err := sql.Open("sqlite3", "path/to/your/database.db")
    if err != nil {
        // Handle error
    }
    defer db.Close()

    // Pass the existing connection to govisual
    app := gin.New()
    visualHandler := govisual.Wrap(
        app,
        govisual.WithSQLiteStorageDB(db, "govisual_requests"),
    )

    // Use visualHandler as your main handler
    http.ListenAndServe(":8080", visualHandler)
}
```

This approach allows you to:

1. Use your preferred SQLite driver
2. Avoid driver registration conflicts
3. Reuse your existing connection

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

### Multi-Storage Example

Example showing different storage backends:

```bash
cd cmd/examples/multistorage
docker-compose up -d  # Start PostgreSQL and Redis
```

Modify the environment variables in `docker-compose.yml` to switch between storage backends.

Visit [Multi-Storage Example](cmd/examples/multistorage/README.md) for detailed instructions.

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
