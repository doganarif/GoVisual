# Using GoVisual with Claude Code

GoVisual's MCP module gives a coding agent eyes into your running app: every captured request — headers, bodies, logs, SQL queries, panic stacks — becomes something the agent can query, replay, and diff.

## Setup

Mount the MCP endpoint next to your wrapped app:

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
	// ... your routes ...

	st := store.WithNotify(store.NewMemory(200))
	app := govisual.Wrap(mux,
		govisual.WithStore(st),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	root := http.NewServeMux()
	root.Handle("/mcp", gvmcp.Handler(st, gvmcp.WithBaseURL("http://localhost:8080")))
	root.Handle("/", app)
	http.ListenAndServe(":8080", root)
}
```

Register it with Claude Code:

```bash
claude mcp add govisual --transport http http://localhost:8080/mcp
```

Sharing the same `store.WithNotify(...)` instance between `Wrap` and the MCP handler is what makes `await_request` push-fast; mounting `/mcp` outside `Wrap` keeps agent traffic out of your captures.

## The debugging loop

A typical session once the tools are connected:

1. **`get_last_error`** — the most recent 4xx/5xx/panic with full context. The usual starting point.
2. **`get_debug_context`** with the request id — request line, headers, bodies, application logs, SQL queries, outbound calls, and panic stack as one readable report.
3. Fix the code, restart the app.
4. **`diff_replay`** with the same id — replays the captured request against the running app and reports `status changed: 500 -> 200` (or that nothing changed).
5. **`save_as_test`** — turn the request that used to fail into a regression test.

For "why is this slow" instead of "why is this broken", start with **`get_stats`** (per-route p50/p95 and error counts), then `get_request` on a slow one — with `WithProfiling(true)` the SQL queries and outbound calls come with durations.

To capture something that hasn't happened yet, call **`await_request`** with filters, then trigger the flow (curl, a test, a browser tool) — the tool returns the matching request the moment it's captured.

## A CLAUDE.md snippet for your project

Drop this into your repo's CLAUDE.md so the agent knows the tools exist and how to use them well:

```markdown
## Runtime debugging

This app runs with GoVisual; the `govisual` MCP server exposes captured HTTP
traffic. When debugging runtime behavior, prefer real captures over guessing:
start with `get_last_error` or `get_stats`, read `get_debug_context` for the
failing request, and verify fixes with `diff_replay` (expect
"status changed" or "no change" — don't claim a fix without it).
`clear_requests` before reproducing gives a clean capture.
```

## Security notes

- The MCP endpoint answers loopback addresses only, unless `gvmcp.WithAllowRemote()` is set. Pair remote access with `gvmcp.WithToken("...")`.
- `replay_request` can change the method, path, headers, and body — but never the destination. Replays always target your app (`WithBaseURL`), so the endpoint is not an SSRF primitive.
- Sensitive headers (Authorization, Cookie, API keys) are redacted at capture time, before anything reaches the store or the agent.
