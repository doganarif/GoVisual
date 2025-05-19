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
