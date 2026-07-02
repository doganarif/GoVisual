package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type awaitArgs struct {
	Method         string `json:"method,omitempty"`
	PathContains   string `json:"path_contains,omitempty"`
	StatusMin      int    `json:"status_min,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

type curlResult struct {
	Command string `json:"command"`
}

type testResult struct {
	Code string `json:"code"`
	Note string `json:"note"`
}

type clearResult struct {
	Cleared bool `json:"cleared"`
}

func registerActionTools(srv *sdk.Server, st store.Store) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "await_request",
		Description: "Block until a request matching the filters (method, path_contains, status_min) is " +
			"captured, then return it in full. Trigger the traffic yourself (curl, a test, a browser tool) " +
			"and use this to catch exactly what happened. timeout_seconds defaults to 30, max 120.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args awaitArgs) (*sdk.CallToolResult, requestDetail, error) {
		timeout := time.Duration(args.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		if timeout > 2*time.Minute {
			timeout = 2 * time.Minute
		}

		var notify <-chan struct{}
		if ns, ok := st.(*store.NotifyingStore); ok {
			ch, cancel := ns.Subscribe()
			defer cancel()
			notify = ch
		}
		// Polling fallback covers stores handed in without notification.
		poll := time.NewTicker(200 * time.Millisecond)
		defer poll.Stop()
		deadline := time.NewTimer(timeout)
		defer deadline.Stop()

		since := time.Now()
		check := func() (*store.RequestLog, bool) {
			for _, l := range st.GetLatest(50) {
				if !l.Timestamp.After(since) {
					continue
				}
				if args.Method != "" && !strings.EqualFold(args.Method, l.Method) {
					continue
				}
				if args.PathContains != "" && !strings.Contains(l.Path, args.PathContains) {
					continue
				}
				if args.StatusMin > 0 && l.StatusCode < args.StatusMin {
					continue
				}
				return l, true
			}
			return nil, false
		}

		for {
			select {
			case <-notify:
			case <-poll.C:
			case <-deadline.C:
				return nil, requestDetail{}, fmt.Errorf("no matching request within %s", timeout)
			case <-ctx.Done():
				return nil, requestDetail{}, ctx.Err()
			}
			if l, ok := check(); ok {
				return nil, detail(l, defaultBodyBytes), nil
			}
		}
	})

	sdk.AddTool(srv, &sdk.Tool{
		Name: "copy_as_curl",
		Description: "Render a captured request as a curl command you can run or share.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, curlResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, curlResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		return nil, curlResult{Command: asCurl(l)}, nil
	})

	sdk.AddTool(srv, &sdk.Tool{
		Name: "save_as_test",
		Description: "Generate a Go httptest regression test from a captured request, asserting the " +
			"captured status code. Paste it into a _test.go file and point it at your handler.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, testResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, testResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		return nil, testResult{
			Code: asTest(l),
			Note: "replace `handler` with your application's http.Handler (usually the mux)",
		}, nil
	})

	sdk.AddTool(srv, &sdk.Tool{
		Name:        "clear_requests",
		Description: "Delete all captured requests. Useful before reproducing an issue for a clean capture.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args emptyArgs) (*sdk.CallToolResult, clearResult, error) {
		if err := st.Clear(); err != nil {
			return nil, clearResult{}, fmt.Errorf("clearing store: %w", err)
		}
		return nil, clearResult{Cleared: true}, nil
	})
}

func asCurl(l *store.RequestLog) string {
	var b strings.Builder
	b.WriteString("curl")
	if l.Method != "GET" {
		fmt.Fprintf(&b, " -X %s", l.Method)
	}
	for k, vs := range l.RequestHeaders {
		for _, v := range vs {
			fmt.Fprintf(&b, " \\\n  -H %s", shellQuote(k+": "+v))
		}
	}
	if l.RequestBody != "" {
		fmt.Fprintf(&b, " \\\n  -d %s", shellQuote(l.RequestBody))
	}
	host := l.Host
	if host == "" {
		host = "localhost:8080"
	}
	url := "http://" + host + l.Path
	if l.Query != "" {
		url += "?" + l.Query
	}
	fmt.Fprintf(&b, " \\\n  %s", shellQuote(url))
	return b.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func asTest(l *store.RequestLog) string {
	var b strings.Builder
	name := testName(l)
	target := l.Path
	if l.Query != "" {
		target += "?" + l.Query
	}

	fmt.Fprintf(&b, "func Test%s(t *testing.T) {\n", name)
	if l.RequestBody != "" {
		fmt.Fprintf(&b, "\tbody := strings.NewReader(%q)\n", l.RequestBody)
		fmt.Fprintf(&b, "\treq := httptest.NewRequest(%q, %q, body)\n", l.Method, target)
	} else {
		fmt.Fprintf(&b, "\treq := httptest.NewRequest(%q, %q, nil)\n", l.Method, target)
	}
	for k, vs := range l.RequestHeaders {
		if k == "Content-Length" || strings.HasPrefix(k, "Accept-Encoding") {
			continue
		}
		for _, v := range vs {
			fmt.Fprintf(&b, "\treq.Header.Set(%q, %q)\n", k, v)
		}
	}
	b.WriteString("\n\trec := httptest.NewRecorder()\n")
	b.WriteString("\thandler.ServeHTTP(rec, req)\n\n")
	fmt.Fprintf(&b, "\tif rec.Code != %d {\n", l.StatusCode)
	fmt.Fprintf(&b, "\t\tt.Fatalf(\"got status %%d, want %d\", rec.Code)\n", l.StatusCode)
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func testName(l *store.RequestLog) string {
	var b strings.Builder
	b.WriteString(strings.ToUpper(l.Method[:1]) + strings.ToLower(l.Method[1:]))
	for _, part := range strings.Split(l.Path, "/") {
		if part == "" {
			continue
		}
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, part)
		if clean == "" {
			continue
		}
		b.WriteString(strings.ToUpper(clean[:1]) + clean[1:])
	}
	return b.String()
}
