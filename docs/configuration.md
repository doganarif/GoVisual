# Configuration Options

All configuration is done through option functions passed to `govisual.Wrap()`.

```go
import "github.com/doganarif/govisual/v2"

handler := govisual.Wrap(
    originalHandler,
    option1,
    option2,
)
```

## Available Options

### Core Options

| Option | Description | Default | Example |
| --- | --- | --- | --- |
| `WithMaxRequests(int)` | Capacity of the default in-memory store. Ignored when a custom store is set via `WithStore`. | 100 | `govisual.WithMaxRequests(500)` |
| `WithDashboardPath(string)` | URL path for the dashboard | `/__viz` | `govisual.WithDashboardPath("/__debug")` |
| `WithRequestBodyLogging(bool)` | Capture request bodies | false | `govisual.WithRequestBodyLogging(true)` |
| `WithResponseBodyLogging(bool)` | Capture response bodies | false | `govisual.WithResponseBodyLogging(true)` |
| `WithIgnorePaths(...string)` | Path patterns to exclude from capture | `["/favicon.ico"]` | `govisual.WithIgnorePaths("/health", "/metrics")` |
| `WithMaxBodyBytes(int)` | Cap on captured body size in bytes. `0` = 1 MiB default, positive = explicit cap, negative = unbounded. | 1 MiB | `govisual.WithMaxBodyBytes(64 << 10)` |
| `WithSampleRate(float64)` | Fraction of requests to capture (0..1). Uncaptured requests pass through untouched. | 1.0 | `govisual.WithSampleRate(0.1)` |
| `WithStore(store.Store)` | Storage backend for captured requests. Omit to use an in-memory store bounded by `WithMaxRequests`. | in-memory | `govisual.WithStore(pg)` |
| `WithErrorHandler(func(error))` | Route capture persistence failures synchronously; callbacks should return promptly and panics are recovered. | standard logger | `govisual.WithErrorHandler(reportError)` |
| `WithShutdownContext(ctx)` | Cancel this context to release storage resources on shutdown. | none | `govisual.WithShutdownContext(ctx)` |

### Dashboard Security

The dashboard is loopback-only by default. `WithAllowRemote` opts out of that restriction; never use it without `WithBasicAuth`, `WithDashboardAuth`, or equivalent authentication in an outer handler.

| Option | Description | Default | Example |
| --- | --- | --- | --- |
| `WithLocalhostOnly()` | Restrict the dashboard to loopback addresses. This is the default; the option exists to make intent explicit. | on | `govisual.WithLocalhostOnly()` |
| `WithAllowRemote()` | Allow non-loopback addresses to reach the dashboard. Must be paired with `WithBasicAuth`, `WithDashboardAuth`, or equivalent outer authentication. | off | `govisual.WithAllowRemote()` |
| `WithBasicAuth(user, pass)` | Protect the dashboard with HTTP Basic Auth (constant-time compare) | off | `govisual.WithBasicAuth("admin", "secret")` |
| `WithDashboardAuth(fn)` | Custom auth function run on every dashboard request; return true to allow | off | `govisual.WithDashboardAuth(myCheck)` |
| `WithReplayEnabled(bool)` | Enable request replay. Replays load a captured request by ID and use a server-pinned destination. | false | `govisual.WithReplayEnabled(true)` |
| `WithReplayBaseURL(string)` | Pin replay to an application origin. Required with `WithAllowRemote()`; a loopback-only dashboard can use its validated origin. | validated loopback dashboard origin | `govisual.WithReplayBaseURL("http://127.0.0.1:8080")` |
| `WithSystemInfo(...string)` | Enable the system-info endpoint; env vars shown only if allowlisted | false | `govisual.WithSystemInfo("GOPATH")` |

### Profiling Options

| Option | Description | Default | Example |
| --- | --- | --- | --- |
| `WithProfiling(bool)` | Enable per-request allocation, GC, goroutine, SQL, and outbound HTTP profiling | false | `govisual.WithProfiling(true)` |
| `WithProfileType(ProfileType)` | Which profiles to collect: `ProfileCPU`, `ProfileMemory`, `ProfileGoroutine`, `ProfileAll` | `ProfileAll` | `govisual.WithProfileType(govisual.ProfileCPU)` |
| `WithProfileThreshold(duration)` | Only keep profiles for requests slower than this | 10ms | `govisual.WithProfileThreshold(50 * time.Millisecond)` |
| `WithMaxProfileMetrics(int)` | Maximum number of profile records to retain | 1000 | `govisual.WithMaxProfileMetrics(500)` |

## Configuration Examples

### Basic Setup

```go
import "github.com/doganarif/govisual/v2"

handler := govisual.Wrap(
    mux,
    govisual.WithMaxRequests(200),
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
    govisual.WithBasicAuth("admin", "secret"),
    govisual.WithReplayEnabled(true),
    govisual.WithReplayBaseURL("http://127.0.0.1:8080"),
    govisual.WithSystemInfo("GOPATH", "GOOS"),
)
```

### PostgreSQL Storage

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/postgres"
)

pg, err := postgres.New(
    "postgres://user:password@localhost:5432/database?sslmode=disable",
    "govisual_requests",
    500, // capacity
)
if err != nil {
    log.Fatal(err)
}

handler := govisual.Wrap(
    mux,
    govisual.WithStore(pg),
)
```

### Sampling Busy Services

```go
// Capture only 10% of requests.
handler := govisual.Wrap(
    mux,
    govisual.WithSampleRate(0.1),
)
```

### Graceful Shutdown

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

handler := govisual.Wrap(
    mux,
    govisual.WithShutdownContext(ctx),
)
```

### Complete Example

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/redis"
)

rdb, err := redis.New("redis://localhost:6379/0", 500, 86400)
if err != nil {
    log.Fatal(err)
}

handler := govisual.Wrap(
    mux,
    govisual.WithStore(rdb),
    govisual.WithDashboardPath("/__debug"),
    govisual.WithRequestBodyLogging(true),
    govisual.WithResponseBodyLogging(true),
    govisual.WithIgnorePaths("/health", "/metrics", "/public/*"),
    govisual.WithSampleRate(0.5),
    govisual.WithBasicAuth("admin", "secret"),
    govisual.WithProfiling(true),
    govisual.WithProfileThreshold(25*time.Millisecond),
    govisual.WithShutdownContext(ctx),
)
```

## Related Documentation

- [Storage Backends](storage-backends.md) - Installing and configuring storage backends
- [API Reference](api-reference.md) - Full API reference
- [Quick Start Guide](quick-start.md) - Getting started with GoVisual
