# GoVisual

A lightweight, zero-configuration HTTP request visualizer and debugger for Go web applications during local development.

### Featured In

- [**Applied Go Weekly**](https://newsletter.appliedgo.net/archive/2025-05-11-tuppers-formula/) — May 2025
- [**GDG Berlin Golang Meetup**](https://www.youtube.com/watch?v=5ZbpWDnYxLM) — Talk: "Building GoVisual: Zero-Config HTTP Debugging in Go" (Jan 2026)
- [**Hacker News**](https://news.ycombinator.com/front?day=2025-05-04) — Show HN front page
- [**Golang Weekly (X)**](https://x.com/golangch/status/1918925327843422248) — 9,100+ impressions

## Features

- **Real-time Request Monitoring**: Visualize HTTP requests passing through your application
- **Request Inspection**: Deep inspection of headers, body, status codes, and timing information
- **Middleware Tracing**: Visualize middleware execution flow and identify performance bottlenecks
- **Zero Configuration**: Drop-in integration with standard Go HTTP handlers
- **OpenTelemetry Integration**: Optional export of telemetry data to OpenTelemetry collectors

## Installation

```bash
go get github.com/doganarif/govisual/v2
```

The core module has no database drivers, no gRPC, nothing you didn't ask for. Storage backends live in separate modules you pull in only when you use them.

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/doganarif/govisual/v2"
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
    govisual.WithMaxBodyBytes(1 << 20),         // Cap captured body size (default 1 MiB)
    govisual.WithIgnorePaths("/health"),        // Paths to ignore

    // Dashboard access (all off by default)
    govisual.WithAllowRemote(),                 // Dashboard is loopback-only by default; opt out here
    govisual.WithBasicAuth("admin", "secret"),  // Protect the dashboard with Basic Auth
    govisual.WithReplayEnabled(true),           // Allow replaying captured requests
    govisual.WithSystemInfo("GOPATH", "HOME"),  // Expose runtime info + allowlisted env vars

    // Profiling
    govisual.WithProfiling(true),               // Per-request CPU/memory profiling
    govisual.WithProfileThreshold(50*time.Millisecond), // Only profile slow requests

    govisual.WithShutdownContext(ctx),          // Release the store when ctx is cancelled

    // Storage (in-memory by default; see Storage Backends below)
    govisual.WithStore(myStore),
)
```

## SQL Query Capture

Wrap your database driver and every query executed with a request's context lands on that request's profile — durations, affected rows, and errors:

```go
sql.Register("postgres+viz", govisual.WrapDriver(&pq.Driver{}))
db, err := sql.Open("postgres+viz", dsn)
```

Run queries with the request context (`db.QueryContext(r.Context(), ...)`) and enable profiling (`WithProfiling(true)`).

## Log Capture

Wrap your slog handler and log lines emitted while handling a request travel with that request:

```go
logger := slog.New(govisual.SlogHandler(slog.NewJSONHandler(os.Stdout, nil)))

// in a handler:
logger.InfoContext(r.Context(), "cache miss", "key", key)
```

Logging still reaches your base handler exactly as before; govisual just keeps a bounded copy (200 lines per request) alongside the captured request.

## OpenTelemetry

Tracing lives in its own module so the OTel SDK and gRPC stay out of builds that don't use them:

```bash
go get github.com/doganarif/govisual/telemetry
```

```go
traced, shutdown, err := telemetry.Wrap(mux, telemetry.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    Endpoint:       "localhost:4317",
    Insecure:       true,
    Exporter:       "otlp", // or "stdout", "noop"
})
if err != nil {
    log.Fatal(err)
}
defer shutdown(context.Background())

handler := govisual.Wrap(traced) // wrapping the traced mux keeps the dashboard out of your traces
```

## Dashboard Security

The dashboard is meant for local development, so everything risky is off unless you opt in:

- **The dashboard only answers loopback addresses by default.** `WithAllowRemote()` opens it up — pair that with auth.
- **Request replay** (`WithReplayEnabled`) is disabled by default — the endpoint makes the server issue outbound HTTP requests.
- **System info** (`WithSystemInfo`) is disabled by default. Environment variables are only exposed if you pass their names explicitly: `WithSystemInfo("GOPATH", "HOME")`.
- `WithBasicAuth(user, pass)` or `WithDashboardAuth(func(*http.Request) bool)` gate every dashboard request.

## Storage Backends

Request logs are kept in memory by default (bounded by `WithMaxRequests`). For persistence, each database backend is its own Go module — installing one is what pulls in its driver:

```bash
go get github.com/doganarif/govisual/store/postgres   # or redis, sqlite, mongodb
```

Construct the store and hand it to `WithStore`:

```go
import (
    "github.com/doganarif/govisual/store/postgres"
    "github.com/doganarif/govisual/v2"
)

st, err := postgres.New(
    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    "govisual_requests", // table, created automatically
    500,                 // capacity
)
if err != nil {
    log.Fatal(err)
}

handler := govisual.Wrap(mux, govisual.WithStore(st))
```

The other backends follow the same shape:

```go
redis.New("redis://localhost:6379/0", 500, 86400)          // capacity, TTL seconds
sqlite.New("requests.db", "govisual_requests", 500)
mongodb.New("mongodb://localhost:27017", "db", "coll", 500)
```

If your application already registers a SQLite driver, reuse your own `*sql.DB` instead of opening a second one:

```go
db, _ := sql.Open("sqlite3", "path/to/your/database.db")
st, err := sqlite.NewWithDB(db, "govisual_requests", 500)
```

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

<!-- arif-signature:start -->

---

Built by [Arif Dogan](https://arif.sh) - production AI and backend engineer.

I help SaaS teams ship production AI features, fast backends, and reliable developer tools.

[Work with me](https://arif.sh/work) | [Book a 30-min intro](https://calendar.superhuman.com/book/11SzDRA4zo8tuYoehO/A3kIl)

<!-- arif-signature:end -->
