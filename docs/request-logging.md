# Request Logging

GoVisual v2 captures request metadata around the `http.Handler` you pass to `govisual.Wrap`. Metadata capture is on by default; request and response bodies are opt-in.

## Captured metadata

Each sampled request records:

- ID and timestamp
- method, host, path, and raw query
- request and response headers
- status code and duration in milliseconds
- request-scoped logs and custom events
- middleware and route trace data when supplied
- profiling data when profiling is enabled and the request meets the configured threshold
- an error and stack trace when the wrapped handler panics

Panics are recorded and then re-raised, so recovery middleware and `net/http` keep their normal behavior.

## Body logging and limits

Body capture is disabled by default:

```go
handler := govisual.Wrap(
	mux,
	govisual.WithRequestBodyLogging(true),
	govisual.WithResponseBodyLogging(true),
)
```

Captured request and response bodies are each limited to 1 MiB by default. Larger bodies are truncated in the stored copy; the application still receives or sends the request normally. Set an explicit cap when needed:

```go
handler := govisual.Wrap(
	mux,
	govisual.WithRequestBodyLogging(true),
	govisual.WithResponseBodyLogging(true),
	govisual.WithMaxBodyBytes(256<<10), // 256 KiB per captured body
)
```

A negative `WithMaxBodyBytes` value disables the cap and is not recommended. Body logging may retain passwords, tokens, personal data, or large payloads; enable it only where that tradeoff is acceptable.

## Header redaction

Headers are captured, but these credential-bearing values are replaced with `[redacted by govisual]` before the request reaches any storage backend:

- `Authorization`
- `Proxy-Authorization`
- `Cookie`
- `Set-Cookie`
- `X-Api-Key`
- `X-Auth-Token`
- `X-Csrf-Token`

Header names remain visible. Apply application-specific redaction before `govisual.Wrap` if your service uses other sensitive headers. Body values are not automatically redacted.

## Sampling and ignored paths

Capture only a fraction of traffic with a rate from 0 to 1:

```go
handler := govisual.Wrap(mux, govisual.WithSampleRate(0.1)) // about 10%
```

Exclude noisy or sensitive paths:

```go
handler := govisual.Wrap(
	mux,
	govisual.WithIgnorePaths(
		"/health",
		"/metrics",
		"/static/*",
	),
)
```

Patterns use Go's `filepath.Match`; a pattern ending in `/` also acts as a prefix. The dashboard path and its API are always ignored to prevent recursive logging. `/favicon.ico` is ignored by default.

## Request-scoped logs and events

Wrap a `slog.Handler`, then log with the incoming request context:

```go
logger := slog.New(govisual.SlogHandler(slog.NewJSONHandler(os.Stdout, nil)))
logger.InfoContext(r.Context(), "cache lookup", "hit", true)
```

Add structured diagnostic events without a logger:

```go
govisual.Event(r.Context(), "cache miss", "key", key, "tier", "redis")
```

Both appear on the request's Logs tab. Records written without the request context still pass to the underlying `slog.Handler`, but GoVisual cannot associate them with a request.

## Stored shape

The core request record is `store.RequestLog`:

```go
type RequestLog struct {
	ID                 string
	Timestamp          time.Time
	Method             string
	Host               string
	Path               string
	RawPath            string // optional encoded path, such as /users/a%2Fb
	Query              string
	RequestHeaders     http.Header
	ResponseHeaders    http.Header
	StatusCode         int
	Duration           int64 // milliseconds
	RequestBody        string
	ResponseBody       string
	Error              string
	MiddlewareTrace    []map[string]interface{}
	RouteTrace         map[string]interface{}
	PerformanceMetrics *PerformanceMetrics
	Logs               []LogEntry
	PanicStack         string
}
```

Durations inside `PerformanceMetrics`, SQL calls, outbound HTTP calls, and nested trace entries are Go `time.Duration` values and therefore encode as nanoseconds in JSON. `RequestLog.Duration`, replay-response duration, and the top-level `MiddlewareTrace` map duration encode as milliseconds; the latter remains in milliseconds for compatibility with persisted v2.0.0 records.

## Storage behavior

Without `WithStore`, records live in an in-memory ring bounded by `WithMaxRequests` (100 by default). PostgreSQL, Redis, MongoDB, and SQLite are separate modules and persist records according to their own configuration. See [Storage Backends](storage-backends.md).

Storage failures do not fail the application request path. Monitor your selected backend separately if durable capture is required.

## Related documentation

- [Dashboard](dashboard.md)
- [Configuration options](configuration.md)
- [Storage backends](storage-backends.md)
- [Middleware tracing](middleware-tracing.md)
