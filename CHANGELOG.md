# Changelog

All notable changes to GoVisual are recorded here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - 2026-07-02

The 2.0 release turns GoVisual into a runtime debugger a coding agent can drive, keeps the core module dependency-free, and hardens every prior rough edge with a full test suite.

### Highlights

- **MCP server for coding agents.** New `github.com/doganarif/govisual/mcp` module serves captured traffic to Claude Code, Cursor, and any other Model Context Protocol client over streamable HTTP. Twelve tools cover the loop: `get_last_error`, `list_requests`, `get_request`, `search_requests`, `get_stats`, `get_debug_context`, `replay_request`, `diff_replay`, `await_request`, `save_as_test`, `copy_as_curl`, `clear_requests`. Responses are token-aware; replays are pinned to a configured base URL, so the endpoint is not an SSRF primitive.
- **Zero-dependency core.** `github.com/doganarif/govisual/v2` requires stdlib plus `google/pprof` and nothing else. Postgres, Redis, SQLite, MongoDB drivers, gRPC and the OpenTelemetry SDK are all separate modules you install only when you use them.
- **Per-request capture.** New public wrappers attach data to the request that caused it:
  - `govisual.WrapDriver` for SQL queries
  - `govisual.WrapTransport` for outbound HTTP calls
  - `govisual.SlogHandler` for slog output (bounded at 200 lines per request)
  - `govisual.Event(ctx, name, kv...)` for application-level markers
  - Handler panics are captured with their full stack, then re-raised so recovery middleware behaves exactly as before.
- **Dashboard rewrite.** Inbox / Errors / Slow filters, per-request drawer with Overview, Headers, Body, Trace, Logs, and Performance tabs, analytics view, agent activity feed, dark mode. Live updates push over SSE.
- **Secure by default.** Dashboard is loopback-only unless `WithAllowRemote()`. Replay and system-info endpoints are opt-in. Sensitive headers (Authorization, Cookie, Set-Cookie, X-Api-Key, X-Auth-Token, X-Csrf-Token) are redacted at capture time.

### New modules

- `github.com/doganarif/govisual/v2` (root, stdlib + pprof)
- `github.com/doganarif/govisual/mcp`
- `github.com/doganarif/govisual/telemetry`
- `github.com/doganarif/govisual/store/postgres`
- `github.com/doganarif/govisual/store/redis`
- `github.com/doganarif/govisual/store/sqlite`
- `github.com/doganarif/govisual/store/mongodb`

### New APIs (root module)

- `govisual.WithStore(store.Store)` replaces the per-backend storage options.
- `govisual.WithActivityLog(*store.ActivityLog)` pairs with `mcp.WithActivityLog` to surface agent tool calls on the dashboard.
- `govisual.WithSampleRate(rate float64)` for capture on chatty services.
- `govisual.WithMaxBodyBytes(n int)` for a configurable body-capture cap (default 1 MiB).
- `govisual.WithAllowRemote()` opts out of the loopback-only default.
- `govisual.WithReplayEnabled(bool)`, `govisual.WithSystemInfo(names ...string)`, `govisual.WithBasicAuth(user, pass)`, `govisual.WithDashboardAuth(func(*http.Request) bool)`.
- `govisual.WithShutdownContext(ctx)` releases resources when the context is cancelled (v1 installed a global signal handler).
- Public store package: `store.Store`, `store.RequestLog`, `store.NewMemory`, `store.WithNotify`, `store.NewActivityLog`.

### Breaking changes

1. **Module path** moves to `github.com/doganarif/govisual/v2`. Update imports.
2. **Storage options moved.** `WithMemoryStorage`, `WithPostgresStorage`, `WithSQLiteStorage`, `WithSQLiteStorageDB`, `WithRedisStorage`, `WithMongoDBStorage` are gone. Install the backend module you want and pass its `New(...)` result to `WithStore(...)`.
3. **OpenTelemetry moved.** `WithOpenTelemetry`, `WithServiceName`, `WithServiceVersion`, `WithOTelEndpoint`, `WithOTelInsecure`, `WithOTelExporter` are gone. Install `github.com/doganarif/govisual/telemetry` and wrap your mux with `telemetry.Wrap(mux, cfg)` before handing it to `govisual.Wrap`.
4. **Dashboard is loopback-only by default.** Existing setups that bind to 0.0.0.0 will keep working for the app, but the dashboard now refuses non-loopback requests unless you opt in with `WithAllowRemote()`.
5. **Library no longer installs signal handlers.** Pass `WithShutdownContext(ctx)` to release resources on shutdown; the host application owns the process lifecycle.
6. **/favicon.ico is on the default ignore list.** Add or replace via `WithIgnorePaths(...)`.

The [Upgrading from v0.2.1](README.md#upgrading-from-v021) section of the README has copy-paste migration snippets.

### Bug fixes since v0.2.1

- `WithDashboardPath` now works. The Preact bundle used to hard-code `/__viz/api` for every fetch and the static assets used absolute URLs, so a custom path loaded a blank shell. The UI now derives its API base at load time, and the route matcher requires a `/` boundary so `/__vizfoo` no longer falls into the dashboard ([#31](https://github.com/doganarif/GoVisual/issues/31)).
- Nested middleware traces were silently dropped. `StartTrace` looked up the wrong parent, so the first level of nesting vanished and every user trace under the profiler's root was lost.
- The profiler produced absurd memory numbers when a GC cycle ran during a request. The MemoryTotalAlloc delta used `MemStats.Alloc` (live heap, which shrinks after GC) and wrapped around uint64. Now uses monotonic `TotalAlloc`.
- Data race on the response writer. Status and body are now read through locked accessors that match the writer's own mutex, so leaked handler goroutines cannot race the middleware's post-handler read.
- Redis backend result ordering. `ZRevRange` order was thrown away by a map iteration in `getLogs`; results now stay aligned with the sorted set, and expired keys prune their own IDs from the index so it does not diverge from the key space.
- Capacity trims for SQL, Mongo, and Redis backends are now expressed as "keep the newest N" instead of "count then delete", so concurrent writes cannot leave the store above capacity.
- `EndProfiling` and `analyzeBottlenecks` now mutate the Metrics struct under the same lock that `RecordSQLQuery` and `RecordHTTPCall` hold, so a leaked recorder cannot race the teardown. `GetMetrics` returns a snapshot.
- SQL backends (Postgres, SQLite) now persist the v2 capture fields (Logs, PanicStack, PerformanceMetrics) via an `extras` JSON/JSONB column. Pre-v2 tables are migrated on open. A new persistence contract test in `store/storetest` prevents future backends from silently dropping these fields.
- Redis and MongoDB backends already persisted the full RequestLog via whole-struct serialization; verified via the new contract test.
- Dashboard SSE now pushes on store `Add` instead of polling every 2 seconds. Subscribing happens before the initial snapshot so an `Add` in that window is not missed.

### Community pull requests that landed

- WithConsoleLogging spirit is now covered by `govisual.SlogHandler` and `govisual.Event`.
- Multi-storage examples updated for the new module layout.
- Documentation folder pass to match the current API.

### Install

```bash
go get github.com/doganarif/govisual/v2
go get github.com/doganarif/govisual/mcp                # optional: MCP server
go get github.com/doganarif/govisual/telemetry          # optional: OpenTelemetry
go get github.com/doganarif/govisual/store/postgres     # or redis, sqlite, mongodb
```

### Thanks

To everyone who filed issues, opened pull requests, or used GoVisual in v1 and reported what did and did not work. Full list of contributors on the [repository page](https://github.com/doganarif/GoVisual/graphs/contributors).

## v0.2.x and earlier

See the [Git history](https://github.com/doganarif/GoVisual/commits/main) and [release notes](https://github.com/doganarif/GoVisual/releases) on GitHub for changes prior to v2.
