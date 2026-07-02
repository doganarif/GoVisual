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
| `WithMaxBodyBytes(int)`         | Cap on captured body size in bytes (0 = 1 MiB default, negative = unbounded) | 1 MiB | `govisual.WithMaxBodyBytes(64 << 10)` |
| `WithShutdownContext(ctx)`      | Release storage and telemetry resources when the context is cancelled | none | `govisual.WithShutdownContext(ctx)` |

### Dashboard Security

| Option                              | Description                                                        | Default  | Example                                        |
| ----------------------------------- | ------------------------------------------------------------------ | -------- | ---------------------------------------------- |
| `WithLocalhostOnly()`               | Only allow dashboard requests from loopback addresses              | off      | `govisual.WithLocalhostOnly()`                 |
| `WithBasicAuth(user, pass)`         | Protect the dashboard with HTTP Basic Auth (constant-time compare) | off      | `govisual.WithBasicAuth("admin", "secret")`    |
| `WithDashboardAuth(fn)`             | Custom auth check run on every dashboard request                   | off      | `govisual.WithDashboardAuth(myCheck)`          |
| `WithReplayEnabled(bool)`           | Enable the request replay endpoint (SSRF primitive — keep it gated) | disabled | `govisual.WithReplayEnabled(true)`             |
| `WithSystemInfo(...string)`         | Enable the system-info endpoint; env vars shown only if allowlisted | disabled | `govisual.WithSystemInfo("GOPATH")`            |

### Profiling Options

| Option                            | Description                                | Default | Example                                              |
| --------------------------------- | ------------------------------------------ | ------- | ---------------------------------------------------- |
| `WithProfiling(bool)`             | Enable per-request CPU/memory profiling    | false   | `govisual.WithProfiling(true)`                       |
| `WithProfileType(type)`           | Which profiles to collect                  | all     | `govisual.WithProfileType(profiling.ProfileCPU)`     |
| `WithProfileThreshold(duration)`  | Only keep profiles for requests slower than this | 10ms | `govisual.WithProfileThreshold(50 * time.Millisecond)` |
| `WithMaxProfileMetrics(int)`      | Maximum number of profile records to keep  | 1000    | `govisual.WithMaxProfileMetrics(500)`                |

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

### Secured Dashboard

```go
handler := govisual.Wrap(
    mux,
    govisual.WithLocalhostOnly(),
    govisual.WithBasicAuth("admin", "secret"),
    govisual.WithReplayEnabled(true),
    govisual.WithSystemInfo("GOPATH", "GOOS"),
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
