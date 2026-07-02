# API Reference

## Core

### `Wrap`

```go
func Wrap(handler http.Handler, opts ...Option) http.Handler
```

Wraps an HTTP handler with GoVisual middleware. Captures requests, serves the dashboard, and applies all options.

**Example:**

```go
import "github.com/doganarif/govisual/v2"

handler := govisual.Wrap(mux, govisual.WithRequestBodyLogging(true))
```

---

## Configuration Options

### `WithMaxRequests`

```go
func WithMaxRequests(max int) Option
```

Sets the capacity of the default in-memory store. Has no effect when a custom store is set via `WithStore`.

**Example:**

```go
govisual.WithMaxRequests(500)
```

---

### `WithDashboardPath`

```go
func WithDashboardPath(path string) Option
```

Sets the URL path for the dashboard. Trailing slashes are stripped.

**Example:**

```go
govisual.WithDashboardPath("/__debug")
```

---

### `WithRequestBodyLogging`

```go
func WithRequestBodyLogging(enabled bool) Option
```

Enables or disables capture of request bodies.

**Example:**

```go
govisual.WithRequestBodyLogging(true)
```

---

### `WithResponseBodyLogging`

```go
func WithResponseBodyLogging(enabled bool) Option
```

Enables or disables capture of response bodies.

**Example:**

```go
govisual.WithResponseBodyLogging(true)
```

---

### `WithIgnorePaths`

```go
func WithIgnorePaths(patterns ...string) Option
```

Adds path patterns to exclude from capture. Patterns follow `filepath.Match` syntax. A trailing slash treats the pattern as a prefix match.

**Example:**

```go
govisual.WithIgnorePaths("/health", "/metrics", "/static/")
```

---

### `WithMaxBodyBytes`

```go
func WithMaxBodyBytes(n int) Option
```

Caps the captured body size per request and response. `0` uses the package default (1 MiB). Positive values set an explicit cap. Negative values disable the cap entirely (not recommended for large downloads).

**Example:**

```go
govisual.WithMaxBodyBytes(64 << 10) // 64 KiB
```

---

### `WithSampleRate`

```go
func WithSampleRate(rate float64) Option
```

Captures only the given fraction of requests (0..1). Values outside this range are clamped. Uncaptured requests pass through with no overhead. Useful when govisual wraps a high-traffic service and full capture would be noisy.

**Example:**

```go
govisual.WithSampleRate(0.1) // capture 10% of requests
```

---

### `WithStore`

```go
func WithStore(s store.Store) Option
```

Sets the storage backend for captured requests. Construct one from a storage module, then pass it here. Without this option, an in-memory store bounded by `WithMaxRequests` is used.

**Example:**

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/postgres"
)

pg, err := postgres.New(connStr, "govisual_requests", 500)
// ...
govisual.WithStore(pg)
```

---

### `WithShutdownContext`

```go
func WithShutdownContext(ctx context.Context) Option
```

Wires govisual's internal cleanup (closing the storage backend) to a caller-provided context. When the context is cancelled, govisual calls `Close()` on the store. This replaces the prior behavior of installing a global signal handler.

One goroutine blocks on `ctx.Done()` for the lifetime of the wrapped handler. Pass a cancellable context in tests to avoid goroutine leaks across test cases.

**Example:**

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

handler := govisual.Wrap(mux, govisual.WithShutdownContext(ctx))
```

---

## Dashboard Security Options

### `WithLocalhostOnly`

```go
func WithLocalhostOnly() Option
```

Restricts the dashboard to requests from loopback addresses. This is the default; the option exists to state intent explicitly.

**Example:**

```go
govisual.WithLocalhostOnly()
```

---

### `WithAllowRemote`

```go
func WithAllowRemote() Option
```

Allows non-loopback addresses to reach the dashboard. Pair with `WithBasicAuth` or `WithDashboardAuth` — an open dashboard exposes every captured request and response body to whoever can reach the port.

**Example:**

```go
govisual.WithAllowRemote()
```

---

### `WithBasicAuth`

```go
func WithBasicAuth(username, password string) Option
```

Protects the dashboard with HTTP Basic Auth using a constant-time comparison.

**Example:**

```go
govisual.WithBasicAuth("admin", "secret")
```

---

### `WithDashboardAuth`

```go
func WithDashboardAuth(fn DashboardAuth) Option
```

Installs a custom authentication function that runs on every dashboard request. Return true to allow access, false to send HTTP 401. Implementations should use constant-time comparisons when checking secrets.

`DashboardAuth` is `func(r *http.Request) bool`.

**Example:**

```go
govisual.WithDashboardAuth(func(r *http.Request) bool {
    return r.Header.Get("X-Debug-Token") == "s3cret"
})
```

---

### `WithReplayEnabled`

```go
func WithReplayEnabled(enabled bool) Option
```

Enables the dashboard's `/api/replay` endpoint. Disabled by default because the endpoint lets the server make arbitrary outbound HTTP requests (an SSRF primitive). Only enable it behind authentication and/or loopback-only access.

**Example:**

```go
govisual.WithReplayEnabled(true)
```

---

### `WithSystemInfo`

```go
func WithSystemInfo(envAllowlist ...string) Option
```

Enables the dashboard's `/api/system-info` endpoint (runtime stats, memory). Environment variables are only surfaced when their names are explicitly passed; with no arguments, no env vars are shown.

**Example:**

```go
govisual.WithSystemInfo("GOPATH", "GOOS")
```

---

## Profiling Options

### `WithProfiling`

```go
func WithProfiling(enabled bool) Option
```

Enables per-request performance profiling. When enabled, each request captures CPU time, memory allocations, goroutine counts, SQL queries (via `WrapDriver`), and outbound HTTP calls (via `WrapTransport`).

**Example:**

```go
govisual.WithProfiling(true)
```

---

### `WithProfileType`

```go
func WithProfileType(profileType ProfileType) Option
```

Selects which profile kinds to collect. Constants are exported from the `govisual` package:

- `govisual.ProfileCPU`
- `govisual.ProfileMemory`
- `govisual.ProfileGoroutine`
- `govisual.ProfileAll` (default)

**Example:**

```go
govisual.WithProfileType(govisual.ProfileCPU)
```

---

### `WithProfileThreshold`

```go
func WithProfileThreshold(threshold time.Duration) Option
```

Only keeps profiles for requests that take longer than this duration. Defaults to 10ms.

**Example:**

```go
govisual.WithProfileThreshold(50 * time.Millisecond)
```

---

### `WithMaxProfileMetrics`

```go
func WithMaxProfileMetrics(max int) Option
```

Sets the maximum number of profile records to retain. Defaults to 1000.

**Example:**

```go
govisual.WithMaxProfileMetrics(500)
```

---

## Instrumentation Helpers

### `WrapDriver`

```go
func WrapDriver(d driver.Driver) driver.Driver
```

Instruments a `database/sql` driver so queries executed with a request's context appear on that request's profile in the dashboard. Register the wrapped driver once and open the database through it. Requires `WithProfiling(true)` and queries must use the `*Context` variants with the request's context.

**Example:**

```go
import (
    "database/sql"
    "github.com/doganarif/govisual/v2"
    "github.com/lib/pq"
)

sql.Register("postgres+viz", govisual.WrapDriver(&pq.Driver{}))
db, err := sql.Open("postgres+viz", dsn)
```

---

### `WrapTransport`

```go
func WrapTransport(rt http.RoundTripper) http.RoundTripper
```

Instruments an `http.RoundTripper` so outbound HTTP calls made while handling a request appear on that request's profile. A nil `rt` wraps `http.DefaultTransport`. Requires `WithProfiling(true)` and outbound requests must carry the incoming request's context.

**Example:**

```go
client := &http.Client{
    Transport: govisual.WrapTransport(nil),
}
resp, err := client.Do(req.WithContext(r.Context()))
```

---

### `SlogHandler`

```go
func SlogHandler(base slog.Handler) slog.Handler
```

Wraps a `slog.Handler` so log records emitted with a request's context are also attached to that request in the dashboard. Records logged outside a captured request pass through to the base handler untouched.

**Example:**

```go
import (
    "log/slog"
    "os"
    "github.com/doganarif/govisual/v2"
)

logger := slog.New(govisual.SlogHandler(slog.NewJSONHandler(os.Stdout, nil)))

// in a handler:
logger.InfoContext(r.Context(), "cache miss", "key", key)
```

---

### `Event`

```go
func Event(ctx context.Context, name string, kv ...any)
```

Annotates the current request with a named application event. The event appears in the request's log timeline and in the middleware trace when profiling is enabled. Key/value pairs follow slog convention. A call outside a captured request is a no-op.

**Example:**

```go
govisual.Event(r.Context(), "cache miss", "key", key, "tier", "redis")
```

---

## store Package

Import path: `github.com/doganarif/govisual/v2/store`

### `Store` Interface

```go
type Store interface {
    Add(log *RequestLog) error
    Get(id string) (*RequestLog, bool)
    GetAll() []*RequestLog
    GetLatest(n int) []*RequestLog
    Clear() error
    Close() error
}
```

All storage backends implement this interface. `Add` returns an error so callers can surface storage failures. `Close` releases any open connections.

---

### `NewMemory`

```go
func NewMemory(capacity int) Store
```

Returns an in-memory ring buffer store. When the buffer is full, the oldest entry is overwritten. A capacity of zero or less defaults to 100.

**Example:**

```go
st := store.NewMemory(500)
```

---

### `WithNotify`

```go
func WithNotify(s Store) *NotifyingStore
```

Wraps any `Store` and signals subscribers after each successful `Add`. Use `Subscribe` to receive those signals.

```go
func (n *NotifyingStore) Subscribe() (<-chan struct{}, func())
```

The channel is buffered (size 1) so bursts coalesce into a single signal. Call the returned cancel function when done to unsubscribe.

**Example:**

```go
ns := store.WithNotify(store.NewMemory(500))

ch, cancel := ns.Subscribe()
defer cancel()

go func() {
    for range ch {
        // new request was captured
    }
}()
```

---

## Storage Backend Constructors

Each backend is a separate Go module. Install with `go get` and pass the result to `WithStore`.

### `postgres.New`

```go
// module: github.com/doganarif/govisual/store/postgres
func New(connStr, tableName string, capacity int) (*Store, error)
```

**Example:**

```go
pg, err := postgres.New("postgres://user:pass@localhost:5432/db?sslmode=disable", "govisual_requests", 500)
```

---

### `redis.New`

```go
// module: github.com/doganarif/govisual/store/redis
func New(connStr string, capacity int, ttlSeconds int) (*Store, error)
```

`ttlSeconds` sets entry expiry; 0 defaults to 86400 (24 hours).

**Example:**

```go
rdb, err := redis.New("redis://localhost:6379/0", 500, 86400)
```

---

### `sqlite.New`

```go
// module: github.com/doganarif/govisual/store/sqlite
func New(dbPath, tableName string, capacity int) (*Store, error)
```

Register a SQLite driver under the name `"sqlite3"` before calling this.

**Example:**

```go
sq, err := sqlite.New("./govisual.db", "govisual_requests", 500)
```

---

### `sqlite.NewWithDB`

```go
func NewWithDB(db *sql.DB, tableName string, capacity int) (*Store, error)
```

Use when your application already holds a `*sql.DB` opened with a SQLite driver. govisual does not call `Close` on a database it did not open.

**Example:**

```go
sq, err := sqlite.NewWithDB(db, "govisual_requests", 500)
```

---

### `mongodb.New`

```go
// module: github.com/doganarif/govisual/store/mongodb
func New(uri, databaseName, collectionName string, capacity int) (*Store, error)
```

**Example:**

```go
mdb, err := mongodb.New("mongodb://localhost:27017", "govisual", "requests", 500)
```

---

## telemetry Module

Import path: `github.com/doganarif/govisual/telemetry`

Install: `go get github.com/doganarif/govisual/telemetry`

The telemetry module is separate so the OTel SDK and gRPC stay out of builds that don't use them.

### `Wrap`

```go
func Wrap(handler http.Handler, cfg Config) (http.Handler, func(context.Context) error, error)
```

Initializes an exporter from `cfg` and returns the handler instrumented with tracing, plus a shutdown function that flushes pending spans. Wrap the application handler before passing to `govisual.Wrap` to keep the dashboard out of your traces.

**Example:**

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/telemetry"
)

traced, shutdown, err := telemetry.Wrap(mux, telemetry.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    Endpoint:       "localhost:4317",
    Insecure:       true,
    Exporter:       "otlp", // "otlp" | "stdout" | "noop"
})
if err != nil {
    log.Fatal(err)
}
defer shutdown(context.Background())

handler := govisual.Wrap(traced)
```

---

### `Config`

```go
type Config struct {
    ServiceName    string
    ServiceVersion string
    Endpoint       string // defaults to "localhost:4317" when Exporter is "otlp"
    Insecure       bool
    Exporter       string // "otlp" (default), "stdout", "noop"
}
```

`Exporter` values:

- `"otlp"` — OTLP gRPC exporter (default)
- `"stdout"` — pretty-prints spans to stdout; useful for debugging
- `"noop"` — discards all spans; useful for benchmarking tracing overhead

---

### `InitTracer`

```go
func InitTracer(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error)
```

Initializes the global OTel trace provider directly. `Wrap` calls this internally; use `InitTracer` when you want to set up the provider yourself and manage the middleware separately via `NewMiddleware`.

---

### `NewMiddleware`

```go
func NewMiddleware(handler http.Handler, serviceName, serviceVersion string) *Middleware
```

Returns an `http.Handler` that starts a span for each request using the already-configured global trace provider. Call after `InitTracer`.

**Example:**

```go
shutdown, err := telemetry.InitTracer(ctx, cfg)
// ...
handler := telemetry.NewMiddleware(mux, "my-service", "1.0.0")
```

---

## Related Documentation

- [Configuration Options](configuration.md) - Options reference with examples
- [Storage Backends](storage-backends.md) - Installing and configuring backends
