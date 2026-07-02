# GoVisual

Runtime HTTP debugger for Go, with an MCP server so your coding agent can debug the app too.

Wrap your handler, get a local dashboard of every request. Point Claude Code or Cursor at the MCP endpoint, and the agent can list requests, replay them against your app, diff the response before and after a fix, and read the logs, SQL queries, and panic stacks that came with each capture.

![GoVisual Dashboard](docs/dashboard.png)

### Featured In

- [**Applied Go Weekly**](https://newsletter.appliedgo.net/archive/2025-05-11-tuppers-formula/), May 2025
- [**GDG Berlin Golang Meetup**](https://www.youtube.com/watch?v=5ZbpWDnYxLM), talk: "Building GoVisual: Zero-Config HTTP Debugging in Go" (Jan 2026)
- [**Hacker News**](https://news.ycombinator.com/front?day=2025-05-04), Show HN front page
- [**Golang Weekly (X)**](https://x.com/golangch/status/1918925327843422248), 9,100+ impressions

## What's new in 2.0

- **MCP server** so coding agents can read, replay, and verify against your running app.
- **Zero-dependency core.** The main module is stdlib plus `google/pprof`. Storage backends (`postgres`, `redis`, `sqlite`, `mongodb`) and `telemetry` are separate modules you pull in only when you use them.
- **Per-request capture** for SQL queries, outbound HTTP, slog output, and panic stacks, all attached to the request that caused them.
- **Dashboard is loopback-only by default.** Replay and system-info endpoints are opt-in.
- **Dark mode**, live push updates, agent activity feed.

Upgrading from v0.2.1? See [Upgrading](#upgrading-from-v021) below.

## Install

```bash
go get github.com/doganarif/govisual/v2
```

The core has no database drivers, no gRPC, no OTel SDK. Add-ons install separately:

```bash
go get github.com/doganarif/govisual/mcp              # MCP server for coding agents
go get github.com/doganarif/govisual/store/postgres   # or redis, sqlite, mongodb
go get github.com/doganarif/govisual/telemetry        # OpenTelemetry export
```

## Quick Start

```go
package main

import (
	"net/http"

	"github.com/doganarif/govisual/v2"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", userHandler)

	handler := govisual.Wrap(mux,
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	http.ListenAndServe(":8080", handler)
}
```

Open the dashboard at http://localhost:8080/__viz.

## Give your coding agent eyes

The `mcp` module serves captured traffic to any MCP client. Claude Code (or Cursor, or a local script) can call twelve tools against your running app: read the most recent error, pull the full crime scene for one request, replay it after you change the code, and diff the response.

```go
package main

import (
	"net/http"

	gvmcp "github.com/doganarif/govisual/mcp"
	"github.com/doganarif/govisual/v2"
	"github.com/doganarif/govisual/v2/store"
)

func main() {
	mux := http.NewServeMux()
	// ...routes...

	st := store.WithNotify(store.NewMemory(200))
	app := govisual.Wrap(mux, govisual.WithStore(st))

	root := http.NewServeMux()
	root.Handle("/mcp", gvmcp.Handler(st, gvmcp.WithBaseURL("http://localhost:8080")))
	root.Handle("/", app)
	http.ListenAndServe(":8080", root)
}
```

Register with Claude Code:

```bash
claude mcp add govisual --transport http http://localhost:8080/mcp
```

**Tools:** `get_last_error`, `list_requests`, `get_request`, `search_requests`, `get_stats`, `get_debug_context`, `replay_request`, `diff_replay`, `await_request`, `save_as_test`, `copy_as_curl`, `clear_requests`.

Responses are token-aware: bounded list sizes, capped body excerpts, sizes reported so the agent knows when to ask for more. The endpoint is loopback-only by default. Replays are pinned to the base URL you configured, so an agent cannot use it to reach arbitrary hosts.

Full guide: [docs/claude-code.md](docs/claude-code.md).

## Per-request capture

### SQL queries

Wrap your database driver. Queries executed with the request's context land on that request's profile with duration, affected rows, and any error.

```go
import (
	"database/sql"

	"github.com/doganarif/govisual/v2"
	_ "github.com/mattn/go-sqlite3" // your driver of choice
)

sql.Register("sqlite3+viz", govisual.WrapDriver(&sqlite3.SQLiteDriver{}))
db, err := sql.Open("sqlite3+viz", "app.db")
// then in a handler:
rows, err := db.QueryContext(r.Context(), "SELECT ...")
```

Enable `govisual.WithProfiling(true)` on the wrap call and the SQL panel shows up on the request drawer's Trace tab.

### Outbound HTTP

Wrap your `http.RoundTripper` and outbound calls made with the inbound request's context are attached to it.

```go
client := &http.Client{Transport: govisual.WrapTransport(nil)}
resp, err := client.Do(req.WithContext(r.Context()))
```

### Application logs (slog)

Wrap your `slog.Handler`. Every record logged with the request's context also lands on that request's Logs tab, capped at 200 lines per request.

```go
logger := slog.New(govisual.SlogHandler(slog.NewJSONHandler(os.Stdout, nil)))

// in a handler:
logger.InfoContext(r.Context(), "cache miss", "key", key)
```

Your base handler still receives every record unchanged.

### Panics

Handler panics land on the captured request with the panic value and full stack, then the panic is re-raised so recovery middleware and net/http behave exactly as they would without govisual in the chain.

### Custom events

```go
govisual.Event(r.Context(), "cache miss", "key", key, "tier", "redis")
```

The event appears on the request's Logs tab and (with profiling on) inside the middleware trace.

## Dashboard

`http://localhost:8080/__viz` by default. Customize with `WithDashboardPath("/__debug")`.

- **Inbox, Errors, Slow**: filter captured requests by status and duration.
- **Request drawer** per request: Overview, Headers, Body, Trace (middleware, SQL, outbound HTTP), Logs, Performance (CPU, memory, GC, flame graph when profiling is on).
- **Analytics**: per-route request counts, p50/p95, error rates.
- **Agents**: recent MCP tool calls with their arguments, so you can watch a coding agent debug the app in real time.
- **Environment**: Go version, GOOS/GOARCH, memory stats, allowlisted env vars.
- **Dark mode**, toggled from the rail; choice persists via `localStorage`.
- **Push updates** (SSE): new requests appear on the dashboard as they land, not on a poll.

## Storage backends

By default requests live in memory bounded by `WithMaxRequests` (100). For persistence, install one of the storage modules and pass its constructor to `WithStore`.

```go
import (
	"github.com/doganarif/govisual/store/postgres"
	"github.com/doganarif/govisual/v2"
)

st, err := postgres.New(
	"postgres://user:pass@localhost:5432/dbname?sslmode=disable",
	"govisual_requests", // table, created and migrated automatically
	500,                 // capacity
)
if err != nil {
	log.Fatal(err)
}

handler := govisual.Wrap(mux, govisual.WithStore(st))
```

Same shape for the others:

```go
redis.New("redis://localhost:6379/0", 500, 86400)         // capacity, TTL seconds
sqlite.New("requests.db", "govisual_requests", 500)
mongodb.New("mongodb://localhost:27017", "db", "coll", 500)
```

If your app already registers a SQLite driver, reuse your own `*sql.DB`:

```go
db, _ := sql.Open("sqlite3", "app.db")
st, err := sqlite.NewWithDB(db, "govisual_requests", 500)
```

The v2 capture fields (logs, panic stacks, performance metrics) round-trip through every backend, including on tables that pre-date v2 (an ALTER TABLE runs on open).

## OpenTelemetry

Tracing lives in its own module so the OTel SDK and gRPC stay out of builds that don't use them.

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

// Wrap the mux (not the final handler) so govisual's dashboard stays untraced.
handler := govisual.Wrap(traced)
```

## Configuration reference

```go
handler := govisual.Wrap(
	mux,

	// Basics
	govisual.WithMaxRequests(100),
	govisual.WithDashboardPath("/__dashboard"),
	govisual.WithRequestBodyLogging(true),
	govisual.WithResponseBodyLogging(true),
	govisual.WithMaxBodyBytes(1<<20),          // default 1 MiB, negative disables the cap
	govisual.WithIgnorePaths("/health"),        // /favicon.ico is ignored by default
	govisual.WithSampleRate(0.1),               // capture 10% of traffic

	// Storage (in-memory by default)
	govisual.WithStore(myStore),

	// Dashboard security
	govisual.WithAllowRemote(),                 // loopback-only by default
	govisual.WithBasicAuth("admin", "s3cret"),
	govisual.WithReplayEnabled(true),           // opt-in, SSRF-checked
	govisual.WithSystemInfo("GOPATH", "HOME"),  // env vars require an explicit allowlist

	// Profiling (feeds the Performance tab and the bottleneck analyzer)
	govisual.WithProfiling(true),
	govisual.WithProfileThreshold(50 * time.Millisecond),

	// Lifecycle
	govisual.WithShutdownContext(ctx),          // release the store when ctx is cancelled

	// Coding-agent integration
	govisual.WithActivityLog(activity),         // pair with mcp.WithActivityLog
)
```

Full option reference: [docs/api-reference.md](docs/api-reference.md). Longer configuration guide: [docs/configuration.md](docs/configuration.md).

## Dashboard security

The dashboard sees every captured request and response, so v2 defaults are conservative:

- **Loopback-only** unless `WithAllowRemote()`. Pair remote access with auth.
- **Request replay** is off unless `WithReplayEnabled(true)`. Enabling it opens `POST /__viz/api/replay`, which makes the server issue an outbound request. Targets that resolve to private IPs are rejected before the call.
- **System info** is off unless `WithSystemInfo(...)`. Environment variables are only exposed if you list them by name.
- **Basic auth** via `WithBasicAuth(user, pass)`, or a custom check via `WithDashboardAuth(func(*http.Request) bool)`.
- **Sensitive headers** (Authorization, Cookie, Set-Cookie, X-Api-Key, X-Auth-Token, X-Csrf-Token) are redacted at capture time. The header name is kept, the value is replaced.

## Upgrading from v0.2.1

Three breaking changes:

1. **Module path.** `github.com/doganarif/govisual` becomes `github.com/doganarif/govisual/v2`. Update imports.
2. **Storage options moved.** `WithMemoryStorage`, `WithPostgresStorage`, `WithSQLiteStorage`, `WithSQLiteStorageDB`, `WithRedisStorage`, `WithMongoDBStorage` are gone. Install the backend module you want and pass the constructor to `WithStore(...)`. See the Storage section above.
3. **OpenTelemetry moved.** `WithOpenTelemetry`, `WithServiceName`, `WithServiceVersion`, `WithOTelEndpoint`, `WithOTelInsecure`, `WithOTelExporter` are gone. Install `github.com/doganarif/govisual/telemetry` and wrap your mux with `telemetry.Wrap(mux, telemetry.Config{...})` before handing it to `govisual.Wrap`.

Also worth noting:

- **The dashboard is loopback-only now.** Add `WithAllowRemote()` if you actually want it reachable over the network, and pair it with auth.
- **Request replay is off** unless you pass `WithReplayEnabled(true)`.
- **System info is off** unless you pass `WithSystemInfo(...)`.
- **The captured RequestLog struct** gained `Logs`, `PanicStack`, and `Host`. If you're reading raw JSON from the API, these are new optional fields; existing consumers keep working.

## Examples

```bash
cd cmd/examples/basic         && go run main.go   # dashboard tour
cd cmd/examples/otel          && go run main.go   # OTel + telemetry module
cd cmd/examples/multistorage  && go run main.go   # postgres / redis / sqlite / mongo
cd cmd/examples/profiling     && go run main.go   # per-request profiling and flame graph
cd cmd/examples/tracing       && go run main.go   # middleware trace, Event, SQL capture
```

## Documentation

- [Quick start](docs/quick-start.md)
- [Configuration options](docs/configuration.md)
- [Storage backends](docs/storage-backends.md)
- [OpenTelemetry](docs/opentelemetry.md)
- [Claude Code / MCP](docs/claude-code.md)
- [API reference](docs/api-reference.md)
- [Troubleshooting](docs/troubleshooting.md)
- [FAQ](docs/faq.md)

## License

MIT.

## Contributing

Pull requests welcome. See [Contributing](docs/contributing.md).

<!-- arif-signature:start -->

---

Built by [Arif Dogan](https://arif.sh), production AI and backend engineer.

I help SaaS teams ship production AI features, fast backends, and reliable developer tools.

[Work with me](https://arif.sh/work) | [Book a 30-min intro](https://calendar.superhuman.com/book/11SzDRA4zo8tuYoehO/A3kIl)

<!-- arif-signature:end -->
