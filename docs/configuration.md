# Configuration Options

GoVisual provides numerous configuration options to customize its behavior to fit your specific needs.

## Setting Configuration Options

All configuration is done through option functions passed to the `govisual.Wrap()` function:

```go
handler := govisual.Wrap(
    originalHandler,
    option1,
    option2,
    // ...more options
)
```

## Available Options

### Core Features

| Option                          | Description                           | Default    | Example                                           |
| ------------------------------- | ------------------------------------- | ---------- | ------------------------------------------------- |
| `WithMaxRequests(int)`          | Number of requests to store in memory | 100        | `govisual.WithMaxRequests(500)`                   |
| `WithDashboardPath(string)`     | URL path for the dashboard            | "/\_\_viz" | `govisual.WithDashboardPath("/__debug")`          |
| `WithRequestBodyLogging(bool)`  | Enable logging of request bodies      | false      | `govisual.WithRequestBodyLogging(true)`           |
| `WithResponseBodyLogging(bool)` | Enable logging of response bodies     | false      | `govisual.WithResponseBodyLogging(true)`          |
| `WithIgnorePaths(...string)`    | Paths to exclude from logging         | []         | `govisual.WithIgnorePaths("/health", "/metrics")` |

### Storage Options

| Option                                                     | Description                     | Default | Example                                                                                                          |
| ---------------------------------------------------------  | ------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------- |
| `WithMemoryStorage()`                                      | Use in-memory storage (default) | N/A     | `govisual.WithMemoryStorage()`                                                                                   |
| `WithPostgresStorage(connStr, tableName)`                  | Use PostgreSQL storage          | N/A     | `govisual.WithPostgresStorage("postgres://user:pass@localhost:5432/db", "govisual_requests")`                    |
| `WithRedisStorage(connStr, ttl)`                           | Use Redis storage               | N/A     | `govisual.WithRedisStorage("redis://localhost:6379/0", 86400)`                                                   |
| `WithMongoDBStorage(uri, databaseName, collectionName)`    | Use MongoDB storage             | N/A     | `govisual.WithMongoDBStorage("mongodb://user:password@localhost:27017/", "your_database", "your_collection")`    |

### OpenTelemetry Options

| Option                       | Description                                | Default          | Example                                            |
| ---------------------------- | ------------------------------------------ | ---------------- | -------------------------------------------------- |
| `WithOpenTelemetry(bool)`    | Enable OpenTelemetry integration           | false            | `govisual.WithOpenTelemetry(true)`                 |
| `WithServiceName(string)`    | Service name for OpenTelemetry             | "govisual"       | `govisual.WithServiceName("my-service")`           |
| `WithServiceVersion(string)` | Service version for OpenTelemetry          | "dev"            | `govisual.WithServiceVersion("1.0.0")`             |
| `WithOTelEndpoint(string)`   | OTLP endpoint for exporting telemetry data | "localhost:4317" | `govisual.WithOTelEndpoint("otel-collector:4317")` |

## Configuration Examples

### Basic Configuration

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(100),
    govisual.WithRequestBodyLogging(true),
    govisual.WithResponseBodyLogging(true),
)
```

### Custom Dashboard Path

```go
handler := govisual.Wrap(
    mux,
    govisual.WithDashboardPath("/__debug"),
)
```

### PostgreSQL Storage

```go
handler := govisual.Wrap(
    mux,
    govisual.WithPostgresStorage(
        "postgres://user:password@localhost:5432/database?sslmode=disable",
        "govisual_requests"
    ),
)
```

### OpenTelemetry Integration

```go
handler := govisual.Wrap(
    mux,
    govisual.WithOpenTelemetry(true),
    govisual.WithServiceName("my-api"),
    govisual.WithServiceVersion("1.2.3"),
    govisual.WithOTelEndpoint("otel-collector:4317"),
)
```

### Complete Example

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(500),
    govisual.WithDashboardPath("/__debug"),
    govisual.WithRequestBodyLogging(true),
    govisual.WithResponseBodyLogging(true),
    govisual.WithIgnorePaths("/health", "/metrics", "/public/*"),
    govisual.WithRedisStorage("redis://localhost:6379/0", 86400),
    govisual.WithOpenTelemetry(true),
    govisual.WithServiceName("user-service"),
    govisual.WithServiceVersion("2.0.1"),
    govisual.WithMongoDBStorage("mongodb://user:password@localhost:27017/", "your_database", "your_collection")
)
```

## Related Documentation

- [Storage Backends](storage-backends.md) - Detailed information about storage options
- [OpenTelemetry Integration](opentelemetry.md) - In-depth guide for using OpenTelemetry
- [Quick Start Guide](quick-start.md) - Getting started with GoVisual
