# API Reference

This document provides a complete reference for the GoVisual API.

## Core Functions

### `Wrap`

```go
func Wrap(handler http.Handler, opts ...Option) http.Handler
```

Wraps an HTTP handler with GoVisual middleware to enable request visualization.

**Parameters:**

- `handler http.Handler`: The original HTTP handler to wrap
- `opts ...Option`: Configuration options

**Returns:**

- `http.Handler`: The wrapped handler

**Example:**

```go
wrapped := govisual.Wrap(originalHandler, options...)
```

## Configuration Options

### Request Handling Options

#### WithMaxRequests

```go
func WithMaxRequests(max int) Option
```

Sets the maximum number of requests to store in memory.

**Parameters:**

- `max int`: Maximum number of requests to store

**Example:**

```go
govisual.WithMaxRequests(500)
```

#### WithDashboardPath

```go
func WithDashboardPath(path string) Option
```

Sets the URL path where the dashboard will be accessible.

**Parameters:**

- `path string`: URL path for the dashboard

**Example:**

```go
govisual.WithDashboardPath("/__debug")
```

#### WithRequestBodyLogging

```go
func WithRequestBodyLogging(enabled bool) Option
```

Enables or disables logging of request bodies.

**Parameters:**

- `enabled bool`: Whether to log request bodies

**Example:**

```go
govisual.WithRequestBodyLogging(true)
```

#### WithResponseBodyLogging

```go
func WithResponseBodyLogging(enabled bool) Option
```

Enables or disables logging of response bodies.

**Parameters:**

- `enabled bool`: Whether to log response bodies

**Example:**

```go
govisual.WithResponseBodyLogging(true)
```

#### WithIgnorePaths

```go
func WithIgnorePaths(patterns ...string) Option
```

Sets path patterns to ignore from request logging.

**Parameters:**

- `patterns ...string`: Path patterns to ignore

**Example:**

```go
govisual.WithIgnorePaths("/health", "/metrics", "/static/*")
```

#### WithMaxBodyBytes

```go
func WithMaxBodyBytes(n int) Option
```

Caps the captured request and response body size. `0` uses the package default (1 MiB), a positive value sets an explicit cap in bytes, and a negative value disables the cap.

**Parameters:**

- `n int`: Body capture cap in bytes

**Example:**

```go
govisual.WithMaxBodyBytes(64 << 10)
```

### Dashboard Security Options

#### WithLocalhostOnly

```go
func WithLocalhostOnly() Option
```

Restricts the dashboard to requests originating from a loopback address.

**Example:**

```go
govisual.WithLocalhostOnly()
```

#### WithBasicAuth

```go
func WithBasicAuth(username, password string) Option
```

Protects the dashboard with HTTP Basic Auth using a constant-time comparison.

**Parameters:**

- `username string`: Expected username
- `password string`: Expected password

**Example:**

```go
govisual.WithBasicAuth("admin", "secret")
```

#### WithDashboardAuth

```go
func WithDashboardAuth(fn DashboardAuth) Option
```

Installs a custom authentication function. It runs on every dashboard request and must return true to allow access.

**Parameters:**

- `fn DashboardAuth`: `func(*http.Request) bool`

**Example:**

```go
govisual.WithDashboardAuth(func(r *http.Request) bool {
    return r.Header.Get("X-Debug-Token") == "s3cret"
})
```

#### WithReplayEnabled

```go
func WithReplayEnabled(enabled bool) Option
```

Enables the dashboard's replay endpoint. Disabled by default: the endpoint makes the server perform outbound HTTP requests (an SSRF primitive), so only enable it behind authentication and/or localhost-only access.

**Parameters:**

- `enabled bool`: Whether replay is allowed

**Example:**

```go
govisual.WithReplayEnabled(true)
```

#### WithSystemInfo

```go
func WithSystemInfo(envAllowlist ...string) Option
```

Enables the dashboard's system-info endpoint. Environment variables are only exposed when their names are passed here; with no names, only memory and runtime info is shown.

**Parameters:**

- `envAllowlist ...string`: Environment variable names to expose

**Example:**

```go
govisual.WithSystemInfo("GOPATH", "GOOS")
```

### Profiling Options

#### WithProfiling

```go
func WithProfiling(enabled bool) Option
```

Enables per-request performance profiling (CPU, memory, goroutines).

**Example:**

```go
govisual.WithProfiling(true)
```

#### WithProfileType

```go
func WithProfileType(profileType profiling.ProfileType) Option
```

Sets which profile types to collect. Defaults to all.

**Example:**

```go
govisual.WithProfileType(profiling.ProfileCPU)
```

#### WithProfileThreshold

```go
func WithProfileThreshold(threshold time.Duration) Option
```

Only keeps profiles for requests slower than the threshold. Defaults to 10ms.

**Example:**

```go
govisual.WithProfileThreshold(50 * time.Millisecond)
```

#### WithMaxProfileMetrics

```go
func WithMaxProfileMetrics(max int) Option
```

Sets the maximum number of profile records to keep. Defaults to 1000.

**Example:**

```go
govisual.WithMaxProfileMetrics(500)
```

### Lifecycle Options

#### WithShutdownContext

```go
func WithShutdownContext(ctx context.Context) Option
```

Wires govisual's internal cleanup (storage backends, OpenTelemetry shutdown) to a caller-provided context: when the context is cancelled, govisual releases its resources. GoVisual does not install signal handlers — shutdown stays under the host application's control.

**Parameters:**

- `ctx context.Context`: Context whose cancellation triggers cleanup

**Example:**

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

handler := govisual.Wrap(mux, govisual.WithShutdownContext(ctx))
```

### Storage Options

#### WithMemoryStorage

```go
func WithMemoryStorage() Option
```

Configures GoVisual to use in-memory storage (default).

**Example:**

```go
govisual.WithMemoryStorage()
```

#### WithPostgresStorage

```go
func WithPostgresStorage(connStr string, tableName string) Option
```

Configures GoVisual to use PostgreSQL storage.

**Parameters:**

- `connStr string`: PostgreSQL connection string
- `tableName string`: Name of the table to use

**Example:**

```go
govisual.WithPostgresStorage(
    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    "govisual_requests"
)
```

#### WithRedisStorage

```go
func WithRedisStorage(connStr string, ttlSeconds int) Option
```

Configures GoVisual to use Redis storage.

**Parameters:**

- `connStr string`: Redis connection string
- `ttlSeconds int`: Time-to-live in seconds

**Example:**

```go
govisual.WithRedisStorage("redis://localhost:6379/0", 86400)
```

### WithMongoDBStorage

```go
func WithMongoDBStorage(uri, databaseName, collectionName string)
```

Configures GoVisual to use MongoDB storage.

**Parameters:**

- `uri string`: MongoDB connection URI
- `databaseName string`: Name of the database to use
- `collectionName string`: Name of the collection to use

**Example:**

```go
govisual.WithMongoDBStorage("mongodb://user:password@localhost:27017", "your_database", "your_collection")
```

### OpenTelemetry Options

#### WithOpenTelemetry

```go
func WithOpenTelemetry(enabled bool) Option
```

Enables or disables OpenTelemetry instrumentation.

**Parameters:**

- `enabled bool`: Whether to enable OpenTelemetry

**Example:**

```go
govisual.WithOpenTelemetry(true)
```

#### WithServiceName

```go
func WithServiceName(name string) Option
```

Sets the service name for OpenTelemetry.

**Parameters:**

- `name string`: Service name

**Example:**

```go
govisual.WithServiceName("my-service")
```

#### WithServiceVersion

```go
func WithServiceVersion(version string) Option
```

Sets the service version for OpenTelemetry.

**Parameters:**

- `version string`: Service version

**Example:**

```go
govisual.WithServiceVersion("1.0.0")
```

#### WithOTelEndpoint

```go
func WithOTelEndpoint(endpoint string) Option
```

Sets the OTLP endpoint for exporting telemetry data.

**Parameters:**

- `endpoint string`: OTLP endpoint

**Example:**

```go
govisual.WithOTelEndpoint("otel-collector:4317")
```

## Related Documentation

- [Configuration Options](configuration.md) - Detailed configuration guide
- [Storage Backends](storage-backends.md) - Storage backend documentation
- [OpenTelemetry Integration](opentelemetry.md) - OpenTelemetry integration guide
