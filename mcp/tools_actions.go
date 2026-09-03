package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sort"
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
	Note    string `json:"note,omitempty"`
}

type testResult struct {
	Code string `json:"code"`
	Note string `json:"note"`
}

type saveTestArgs struct {
	ID             string `json:"id"`
	ExpectedStatus *int   `json:"expected_status,omitempty"`
}

type clearResult struct {
	Cleared bool `json:"cleared"`
}

const maxGeneratedBytes = 256 << 10

func registerActionTools(srv *sdk.Server, st store.Store, cfg *config) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "await_request",
		Description: "Block until a request matching the filters (method, path_contains, status_min) is " +
			"captured, then return it in full. Trigger the traffic yourself (curl, a test, a browser tool) " +
			"and use this to catch exactly what happened. timeout_seconds defaults to 30, max 120.",
	}, recorded(cfg, "await_request", false, func(ctx context.Context, req *sdk.CallToolRequest, args awaitArgs) (*sdk.CallToolResult, requestDetail, error) {
		// Snapshot IDs rather than timestamps. A request may begin before this
		// call but only reach the store afterward, and bursts can share a clock
		// tick. Such requests must still be eligible.
		seen := make(map[string]struct{})
		for _, l := range st.GetAll() {
			seen[l.ID] = struct{}{}
		}
		timeout := time.Duration(args.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		if timeout > 2*time.Minute {
			timeout = 2 * time.Minute
		}

		var notify <-chan struct{}
		var poll <-chan time.Time
		var pollTicker *time.Ticker
		if ns, ok := st.(*store.NotifyingStore); ok {
			ch, cancel := ns.Subscribe()
			defer cancel()
			notify = ch
		} else {
			// Stores without notifications need a bounded polling fallback.
			pollTicker = time.NewTicker(200 * time.Millisecond)
			defer pollTicker.Stop()
			poll = pollTicker.C
		}
		deadline := time.NewTimer(timeout)
		defer deadline.Stop()

		check := func() (*store.RequestLog, bool) {
			// The Store API has no cursor. Inspect all IDs so polling stores and
			// coalesced notification bursts cannot bury a matching request.
			logs := st.GetAll()
			for _, l := range logs {
				if _, ok := seen[l.ID]; ok {
					continue
				}
				seen[l.ID] = struct{}{}
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
			if l, ok := check(); ok {
				return nil, detail(l, defaultBodyBytes), nil
			}
			select {
			case <-notify:
			case <-poll:
			case <-deadline.C:
				return nil, requestDetail{}, fmt.Errorf("no matching request within %s", timeout)
			case <-ctx.Done():
				return nil, requestDetail{}, ctx.Err()
			}
		}
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name:        "copy_as_curl",
		Description: "Render a captured request as a curl command you can run or share.",
	}, recorded(cfg, "copy_as_curl", false, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, curlResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, curlResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		if strings.HasSuffix(l.RequestBody, captureTruncationMarker) {
			return nil, curlResult{}, fmt.Errorf("captured request body is truncated; copy would not reproduce the original request")
		}
		if len(l.RequestBody) > maxBodyBytes {
			return nil, curlResult{}, fmt.Errorf("captured request body is too large to render safely (%d bytes; maximum %d)", len(l.RequestBody), maxBodyBytes)
		}
		baseURL, err := replayBaseURL(cfg, l)
		if err != nil {
			return nil, curlResult{}, err
		}
		target, err := replayTarget(baseURL, l.Path, l.RawPath, l.Query, "")
		if err != nil {
			return nil, curlResult{}, err
		}
		if err := validateGeneratedInput(l, target); err != nil {
			return nil, curlResult{}, err
		}
		note := "the destination is pinned by WithBaseURL; review it before running"
		if omitted := omittedRedactedHeaders(l.RequestHeaders, nil); len(omitted) > 0 {
			note += "; supply omitted redacted headers manually: " + strings.Join(omitted, ", ")
		}
		command := asCurl(l, target)
		if len(command) > maxGeneratedBytes {
			return nil, curlResult{}, fmt.Errorf("generated curl command exceeds %d bytes", maxGeneratedBytes)
		}
		return nil, curlResult{Command: command, Note: note}, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "save_as_test",
		Description: "Generate a Go httptest regression test from a captured request, asserting the " +
			"captured status code or optional expected_status. Paste it into a _test.go file and point it at your handler.",
	}, recorded(cfg, "save_as_test", false, func(ctx context.Context, req *sdk.CallToolRequest, args saveTestArgs) (*sdk.CallToolResult, testResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, testResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		if strings.HasSuffix(l.RequestBody, captureTruncationMarker) {
			return nil, testResult{}, fmt.Errorf("captured request body is truncated; generated test would not reproduce the original request")
		}
		if len(l.RequestBody) > maxBodyBytes {
			return nil, testResult{}, fmt.Errorf("captured request body is too large to render safely (%d bytes; maximum %d)", len(l.RequestBody), maxBodyBytes)
		}
		target := l.RawPath
		if target == "" {
			target = l.Path
		}
		if l.Query != "" {
			target += "?" + l.Query
		}
		if err := validateGeneratedInput(l, target); err != nil {
			return nil, testResult{}, err
		}
		expectedStatus := l.StatusCode
		if args.ExpectedStatus != nil {
			if *args.ExpectedStatus < 100 || *args.ExpectedStatus > 999 {
				return nil, testResult{}, fmt.Errorf("expected_status must be between 100 and 999")
			}
			expectedStatus = *args.ExpectedStatus
		}
		note := "replace `handler` with your application's http.Handler (usually the mux)"
		if omitted := omittedRedactedHeaders(l.RequestHeaders, nil); len(omitted) > 0 {
			note += "; supply omitted redacted headers manually: " + strings.Join(omitted, ", ")
		}
		code := asTest(l, expectedStatus)
		if len(code) > maxGeneratedBytes {
			return nil, testResult{}, fmt.Errorf("generated test exceeds %d bytes", maxGeneratedBytes)
		}
		return nil, testResult{
			Code: code,
			Note: note,
		}, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name:        "clear_requests",
		Description: "Delete all captured requests. Useful before reproducing an issue for a clean capture.",
	}, recorded(cfg, "clear_requests", true, func(ctx context.Context, req *sdk.CallToolRequest, args emptyArgs) (*sdk.CallToolResult, clearResult, error) {
		if err := st.Clear(); err != nil {
			return nil, clearResult{}, fmt.Errorf("clearing store: %w", err)
		}
		return nil, clearResult{Cleared: true}, nil
	}))
}

func asCurl(l *store.RequestLog, target string) string {
	var b strings.Builder
	b.WriteString("curl")
	fmt.Fprintf(&b, " --request %s", shellQuote(l.Method))
	headers := make(http.Header)
	copyReplayHeaders(headers, l.RequestHeaders, nil, nil)
	for _, k := range sortedHeaderKeys(headers) {
		vs := headers[k]
		for _, v := range vs {
			fmt.Fprintf(&b, " \\\n  -H %s", shellQuote(k+": "+v))
		}
	}
	if l.RequestBody != "" {
		fmt.Fprintf(&b, " \\\n  --data-raw %s", shellQuote(l.RequestBody))
	}
	fmt.Fprintf(&b, " \\\n  %s", shellQuote(target))
	return b.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func validateGeneratedInput(l *store.RequestLog, target string) error {
	remaining := maxGeneratedBytes
	consume := func(value string) bool {
		if len(value) > remaining {
			return false
		}
		remaining -= len(value)
		return true
	}
	for _, value := range []string{l.Method, l.Host, target, l.RequestBody} {
		if !consume(value) {
			return fmt.Errorf("captured request is too large to render safely (maximum %d bytes)", maxGeneratedBytes)
		}
	}
	for key, values := range l.RequestHeaders {
		if len(key) > maxHeaderBytes || !consume(key) {
			return fmt.Errorf("captured request headers are too large to render safely")
		}
		for _, value := range values {
			if len(value) > maxTextBytes || !consume(value) {
				return fmt.Errorf("captured request headers are too large to render safely")
			}
		}
	}
	return nil
}

func asTest(l *store.RequestLog, expectedStatus int) string {
	var b strings.Builder
	name := testName(l)
	target := l.RawPath
	if target == "" {
		target = l.Path
	}
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
	if l.Host != "" {
		fmt.Fprintf(&b, "\treq.Host = %q\n", l.Host)
	}
	headers := make(http.Header)
	copyReplayHeaders(headers, l.RequestHeaders, nil, nil)
	for _, k := range sortedHeaderKeys(headers) {
		vs := headers[k]
		for _, v := range vs {
			fmt.Fprintf(&b, "\treq.Header.Add(%q, %q)\n", k, v)
		}
	}
	b.WriteString("\n\trec := httptest.NewRecorder()\n")
	b.WriteString("\thandler.ServeHTTP(rec, req)\n\n")
	fmt.Fprintf(&b, "\tif rec.Code != %d {\n", expectedStatus)
	fmt.Fprintf(&b, "\t\tt.Fatalf(\"got status %%d, want %d\", rec.Code)\n", expectedStatus)
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func sortedHeaderKeys(headers http.Header) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func testName(l *store.RequestLog) string {
	var b strings.Builder
	b.WriteString(identifierPart(l.Method, "Request"))
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

func identifierPart(value, fallback string) string {
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToLower(value))
	if clean == "" {
		return fallback
	}
	if clean[0] >= '0' && clean[0] <= '9' {
		clean = "M" + clean
	}
	return strings.ToUpper(clean[:1]) + clean[1:]
}
