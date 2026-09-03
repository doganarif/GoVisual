package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func seedStore(t *testing.T) store.Store {
	t.Helper()
	st := store.NewMemory(50)
	base := time.Now().Add(-time.Minute)
	st.Add(&store.RequestLog{
		ID: "ok-1", Timestamp: base, Method: "GET", Path: "/api/users",
		StatusCode: 200, Duration: 12, Host: "localhost:8080",
	})
	st.Add(&store.RequestLog{
		ID: "ok-2", Timestamp: base.Add(time.Second), Method: "POST", Path: "/api/users",
		StatusCode: 201, Duration: 30, RequestBody: `{"name":"alice"}`, Host: "localhost:8080",
	})
	st.Add(&store.RequestLog{
		ID: "bad-1", Timestamp: base.Add(2 * time.Second), Method: "GET", Path: "/api/orders",
		StatusCode: 500, Duration: 87, Error: "panic: nil map write", PanicStack: "goroutine 1 [running]: ...",
		Logs: []store.LogEntry{{Time: base, Level: "ERROR", Message: "order lookup failed"}},
		Host: "localhost:8080",
	})
	return st
}

func connect(t *testing.T, st store.Store, cfg *config) *sdk.ClientSession {
	t.Helper()
	if cfg == nil {
		cfg = &config{}
	}
	srv := newServer(st, cfg)
	ct, srvT := sdk.NewInMemoryTransports()

	ctx := t.Context()
	if _, err := srv.Connect(ctx, srvT, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func call(t *testing.T, s *sdk.ClientSession, tool string, args any) map[string]any {
	t.Helper()
	res, err := s.CallTool(t.Context(), &sdk.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	if res.IsError {
		t.Fatalf("%s returned tool error: %+v", tool, res.Content)
	}
	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		// Some tools return arrays or strings; wrap them.
		return map[string]any{"value": res.StructuredContent}
	}
	return out
}

func TestGetLastError(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "get_last_error", struct{}{})
	if out["id"] != "bad-1" {
		t.Fatalf("expected bad-1, got %v", out["id"])
	}
	if out["error"] != "panic: nil map write" {
		t.Fatalf("expected panic error, got %v", out["error"])
	}
}

func TestListRequestsErrorsOnly(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "list_requests", map[string]any{"errors_only": true})
	reqs := out["requests"].([]any)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 error request, got %d", len(reqs))
	}
}

func TestSearchRequests(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "search_requests", map[string]any{"query": "alice"})
	reqs := out["requests"].([]any)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 match, got %d", len(reqs))
	}
}

func TestGetDebugContext(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "get_debug_context", map[string]any{"id": "bad-1"})
	report, _ := out["report"].(string)
	for _, want := range []string{"GET /api/orders", "panic: nil map write", "order lookup failed", "Panic stack"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report missing %q:\n%s", want, report)
		}
	}
}

func TestDiffReplayDetectsFix(t *testing.T) {
	// The "fixed" app now returns 200 where the capture saw 500.
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"orders":[]}`))
	}))
	defer app.Close()

	st := seedStore(t)
	u, _ := url.Parse(app.URL)
	s := connect(t, st, &config{baseURL: "http://" + u.Host})

	out := call(t, s, "diff_replay", map[string]any{"id": "bad-1"})
	if out["status_changed"] != true {
		t.Fatalf("expected status change, got %+v", out)
	}
	if out["replay_status"].(float64) != 200 {
		t.Fatalf("expected replay status 200, got %v", out["replay_status"])
	}
	summary := out["summary"].(string)
	if !strings.Contains(summary, "500 -> 200") {
		t.Fatalf("summary = %q", summary)
	}
}

func TestReplayOverrides(t *testing.T) {
	var got *http.Request
	var gotBody string
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		b := make([]byte, 1024)
		n, _ := r.Body.Read(b)
		gotBody = string(b[:n])
		w.WriteHeader(http.StatusAccepted)
	}))
	defer app.Close()

	st := seedStore(t)
	u, _ := url.Parse(app.URL)
	s := connect(t, st, &config{baseURL: "http://" + u.Host})

	out := call(t, s, "replay_request", map[string]any{
		"id":      "ok-2",
		"path":    "/api/users/7",
		"headers": map[string]string{"X-Debug": "1"},
		"body":    `{"name":"bob"}`,
	})
	if out["status"].(float64) != 202 {
		t.Fatalf("expected 202, got %v", out["status"])
	}
	if got.URL.Path != "/api/users/7" || got.Header.Get("X-Debug") != "1" || gotBody != `{"name":"bob"}` {
		t.Fatalf("overrides not applied: %s %v body=%q", got.URL.Path, got.Header, gotBody)
	}
	if got.Header.Get("X-Govisual-Replay") != "1" {
		t.Fatal("replay marker header missing")
	}
}

func TestReplayRequiresPinnedBaseAndSanitizesRequest(t *testing.T) {
	type received struct {
		escapedPath string
		query       string
		body        string
		headers     http.Header
	}
	got := make(chan received, 1)
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got <- received{
			escapedPath: r.URL.EscapedPath(),
			query:       r.URL.RawQuery,
			body:        string(body),
			headers:     r.Header.Clone(),
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer app.Close()

	captured := &store.RequestLog{
		ID:          "hostile-host",
		Method:      http.MethodPost,
		Host:        "169.254.169.254",
		Path:        "/original",
		RequestBody: "cannot-clear",
		RequestHeaders: http.Header{
			"Authorization":       []string{redactedHeaderValue},
			"Proxy-Authorization": []string{redactedHeaderValue},
			"Accept-Encoding":     []string{"br"},
			"Connection":          []string{"X-Hop"},
			"X-Hop":               []string{"remove"},
			"X-Override-Hop":      []string{redactedHeaderValue},
			"X-Remove":            []string{"remove"},
			"X-Safe-Secret":       []string{redactedHeaderValue},
		},
	}
	if _, err := replay(t.Context(), &config{}, captured, replayArgs{}); err == nil || !strings.Contains(err.Error(), "configure WithBaseURL") {
		t.Fatalf("replay without fixed base error = %v", err)
	}

	emptyBody := ""
	result, err := replay(t.Context(), &config{baseURL: app.URL}, captured, replayArgs{
		Path:          "/users/a%2Fb?x=1",
		Body:          &emptyBody,
		RemoveHeaders: []string{"X-Remove"},
		Headers: map[string]string{
			"Connection":          "X-Override-Hop",
			"Proxy-Authorization": "Basic replacement",
			"X-Override-Hop":      "replacement",
			"X-Api-Key":           redactedHeaderValue,
			"X-Debug":             "yes",
			"X-Safe-Secret":       "replacement",
		},
	})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if result.Status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", result.Status, http.StatusNoContent)
	}
	wantOmitted := []string{"Authorization", "Proxy-Authorization", "X-Override-Hop"}
	if !slices.Equal(result.OmittedHeaders, wantOmitted) {
		t.Fatalf("omitted headers = %v, want %v", result.OmittedHeaders, wantOmitted)
	}
	select {
	case request := <-got:
		if request.escapedPath != "/users/a%2Fb" || request.query != "x=1" || request.body != "" {
			t.Fatalf("request shape = %+v", request)
		}
		for _, name := range []string{"Authorization", "Connection", "X-Hop", "X-Override-Hop", "X-Api-Key", "X-Remove"} {
			if value := request.headers.Get(name); value != "" {
				t.Fatalf("blocked header %s = %q", name, value)
			}
		}
		if request.headers.Get("X-Debug") != "yes" {
			t.Fatalf("safe override missing: %v", request.headers)
		}
		if request.headers.Get("X-Safe-Secret") != "replacement" {
			t.Fatalf("safe credential override missing: %v", request.headers)
		}
		if request.headers.Get("Accept-Encoding") != "br" {
			t.Fatalf("end-to-end Accept-Encoding was not preserved: %v", request.headers)
		}
	case <-time.After(time.Second):
		t.Fatal("pinned application did not receive replay")
	}
}

func TestReplayBaseNeverFallsBackToCapturedHost(t *testing.T) {
	for _, host := range []string{"localhost:8080", "127.0.0.1:8080", "169.254.169.254"} {
		if _, err := replayBaseURL(&config{}, &store.RequestLog{Host: host}); err == nil {
			t.Fatalf("captured Host %q was accepted without WithBaseURL", host)
		}
	}
}

func TestReplayTargetPreservesCapturedEscapedPath(t *testing.T) {
	target, err := replayTarget("http://127.0.0.1:8080/base", "/users/a/b", "/users/a%2Fb", "x=1", "")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/base/users/a/b" || parsed.EscapedPath() != "/base/users/a%2Fb" || parsed.RawQuery != "x=1" {
		t.Fatalf("target = %q (path=%q escaped=%q query=%q)", target, parsed.Path, parsed.EscapedPath(), parsed.RawQuery)
	}
}

func TestReplayRejectsAuthorityShapedPaths(t *testing.T) {
	for _, path := range []string{"@attacker.example/x", "//attacker.example/x", "http://attacker.example/x"} {
		t.Run(path, func(t *testing.T) {
			if _, err := replayTarget("http://127.0.0.1:8080", "/captured", "", "", path); err == nil {
				t.Fatalf("replayTarget accepted %q", path)
			}
		})
	}
}

func TestHandlerGates(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	h := Handler(store.NewMemory(10))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.9:1234"
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("remote request: got %d, want 403", rec.Code)
	}

	h = Handler(store.NewMemory(10), WithToken("s3cret"))
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Host = "localhost:8080"
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: got %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Host = "localhost:8080"
	req.Header.Set("Authorization", "Bearer s3cret")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorized initialize: got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerRejectsUntrustedHostAndCrossOrigin(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	h := Handler(store.NewMemory(10))

	tests := []struct {
		name   string
		host   string
		origin string
		want   int
	}{
		{name: "same origin localhost", host: "localhost:8080", origin: "http://localhost:8080", want: http.StatusOK},
		{name: "same origin ipv6", host: "[::1]:8080", origin: "http://[::1]:8080", want: http.StatusOK},
		{name: "dns rebinding host", host: "attacker.example", want: http.StatusForbidden},
		{name: "localhost suffix trick", host: "localhost.attacker.example", want: http.StatusForbidden},
		{name: "cross origin", host: "localhost:8080", origin: "http://attacker.example", want: http.StatusForbidden},
		{name: "different port", host: "localhost:8080", origin: "http://localhost:8081", want: http.StatusForbidden},
		{name: "opaque origin", host: "localhost:8080", origin: "null", want: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			req.RemoteAddr = "127.0.0.1:1234"
			req.Host = tt.host
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("got %d, want %d: %s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAwaitRequest(t *testing.T) {
	st := store.WithNotify(store.NewMemory(50))
	s := connect(t, st, nil)

	go func() {
		time.Sleep(100 * time.Millisecond)
		st.Add(&store.RequestLog{
			ID: "late-1", Timestamp: time.Now().Add(-time.Hour), Method: "POST", Path: "/api/orders",
			StatusCode: 500, Host: "localhost:8080",
		})
	}()

	out := call(t, s, "await_request", map[string]any{
		"path_contains": "/api/orders", "status_min": 500, "timeout_seconds": 5,
	})
	if out["id"] != "late-1" {
		t.Fatalf("expected late-1, got %v", out["id"])
	}
}

func TestAwaitRequestDoesNotMissNonNotifyingBurst(t *testing.T) {
	st := store.NewMemory(200)
	s := connect(t, st, nil)

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = st.Add(&store.RequestLog{ID: "burst-match", Method: "POST", Path: "/match", StatusCode: 500})
		for i := 0; i < maxListLimit; i++ {
			_ = st.Add(&store.RequestLog{ID: fmt.Sprintf("burst-%03d", i), Method: "GET", Path: "/noise", StatusCode: 200})
		}
	}()

	out := call(t, s, "await_request", map[string]any{
		"path_contains": "/match", "timeout_seconds": 5,
	})
	if out["id"] != "burst-match" {
		t.Fatalf("expected burst-match, got %v", out["id"])
	}
}

func TestCopyAsCurl(t *testing.T) {
	s := connect(t, seedStore(t), &config{baseURL: "http://localhost:8080"})
	out := call(t, s, "copy_as_curl", map[string]any{"id": "ok-2"})
	cmd := out["command"].(string)
	for _, want := range []string{"curl", "--request 'POST'", "--data-raw", `{"name":"alice"}`, "http://localhost:8080/api/users"} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("curl command missing %q: %s", want, cmd)
		}
	}
}

func TestCopyAsCurlQuotesMethodAndDisablesAtFileExpansion(t *testing.T) {
	log := &store.RequestLog{
		Method:      "GET|sh",
		Host:        "localhost:8080",
		Path:        "/items",
		RequestBody: "@/etc/passwd",
	}
	command := asCurl(log, "http://localhost:8080/items")
	for _, want := range []string{"--request 'GET|sh'", "--data-raw '@/etc/passwd'"} {
		if !strings.Contains(command, want) {
			t.Fatalf("safe curl command missing %q: %s", want, command)
		}
	}
}

func TestGeneratedTestNameSanitizesCustomMethod(t *testing.T) {
	name := testName(&store.RequestLog{Method: "M-SEARCH|`", Path: "/"})
	if name != "Msearch" {
		t.Fatalf("test name = %q, want Msearch", name)
	}
}

func TestGeneratedArtifactsRejectOversizedHeaders(t *testing.T) {
	log := &store.RequestLog{
		Method: http.MethodGet,
		Path:   "/",
		RequestHeaders: http.Header{
			strings.Repeat("X", maxHeaderBytes+1): []string{"value"},
		},
	}
	if err := validateGeneratedInput(log, "http://localhost:8080/"); err == nil {
		t.Fatal("oversized header name was accepted")
	}
}

func TestActivityArgumentsOmitBodiesAndCredentials(t *testing.T) {
	body := "super-secret-body"
	got := summarizeArgs(replayArgs{
		ID:   "request-1",
		Body: &body,
		Headers: map[string]string{
			"Authorization":       "Bearer secret",
			"Proxy-Authorization": "Basic secret",
			"Cookie":              "session=secret",
			"X-Debug":             "safe",
		},
	})
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, secret := range []string{"super-secret-body", "Bearer secret", "Basic secret", "session=secret"} {
		if strings.Contains(text, secret) {
			t.Fatalf("activity arguments leaked %q: %s", secret, text)
		}
	}
	if !strings.Contains(text, "X-Debug") || !strings.Contains(text, "safe") {
		t.Fatalf("safe diagnostic header was removed: %s", text)
	}
}

func TestBodyChangedPreservesLargeJSONNumbers(t *testing.T) {
	if !bodyChanged(`{"id":9007199254740992}`, `{"id":9007199254740993}`) {
		t.Fatal("distinct large JSON integers compared equal")
	}
}

func TestSaveAsTest(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "save_as_test", map[string]any{"id": "bad-1"})
	code := out["code"].(string)
	for _, want := range []string{"func TestGetApiOrders(t *testing.T)", `httptest.NewRequest("GET", "/api/orders", nil)`, "rec.Code != 500"} {
		if !strings.Contains(code, want) {
			t.Fatalf("generated test missing %q:\n%s", want, code)
		}
	}
}

func TestSaveAsTestExpectedStatusOverride(t *testing.T) {
	s := connect(t, seedStore(t), nil)
	out := call(t, s, "save_as_test", map[string]any{"id": "bad-1", "expected_status": 200})
	code := out["code"].(string)
	if !strings.Contains(code, "rec.Code != 200") || strings.Contains(code, "rec.Code != 500") {
		t.Fatalf("generated test did not use expected_status override:\n%s", code)
	}
}

func TestSaveAsTestPreservesHostAndHeaderValuesSafely(t *testing.T) {
	st := store.NewMemory(1)
	if err := st.Add(&store.RequestLog{
		ID:          "generated-test",
		Method:      http.MethodPost,
		Host:        "tenant.example.test",
		Path:        "/submit",
		StatusCode:  http.StatusCreated,
		RequestBody: "complete body",
		RequestHeaders: http.Header{
			"X-Multi":       []string{"one", "two"},
			"Authorization": []string{redactedHeaderValue},
		},
	}); err != nil {
		t.Fatal(err)
	}
	s := connect(t, st, nil)
	out := call(t, s, "save_as_test", map[string]any{"id": "generated-test"})
	code := out["code"].(string)
	for _, want := range []string{
		`req.Host = "tenant.example.test"`,
		`req.Header.Add("X-Multi", "one")`,
		`req.Header.Add("X-Multi", "two")`,
	} {
		if !strings.Contains(code, want) {
			t.Fatalf("generated test missing %q:\n%s", want, code)
		}
	}
	if strings.Contains(code, redactedHeaderValue) || strings.Contains(code, "Authorization") {
		t.Fatalf("generated test forwarded a redacted credential:\n%s", code)
	}
	if note := out["note"].(string); !strings.Contains(note, "Authorization") {
		t.Fatalf("generated test note = %q, want omitted-header warning", note)
	}
}

func TestReadToolLimitsAreCapped(t *testing.T) {
	st := store.NewMemory(maxListLimit + 50)
	longBody := strings.Repeat("x", maxBodyBytes+100)
	for i := 0; i < maxListLimit+50; i++ {
		st.Add(&store.RequestLog{
			ID:           fmt.Sprintf("request-%03d", i),
			Timestamp:    time.Now().Add(time.Duration(i) * time.Millisecond),
			Method:       http.MethodPost,
			Path:         "/needle",
			RequestBody:  longBody,
			ResponseBody: longBody,
		})
	}
	s := connect(t, st, nil)

	for _, tool := range []string{"list_requests", "search_requests"} {
		args := map[string]any{"limit": maxListLimit * 100}
		if tool == "search_requests" {
			args["query"] = "needle"
		}
		out := call(t, s, tool, args)
		requests := out["requests"].([]any)
		if len(requests) != maxListLimit {
			t.Fatalf("%s returned %d requests, want hard cap %d", tool, len(requests), maxListLimit)
		}
	}

	out := call(t, s, "get_request", map[string]any{
		"id":             fmt.Sprintf("request-%03d", maxListLimit+49),
		"max_body_bytes": maxBodyBytes * 100,
	})
	requestBody := out["request_body"].(string)
	if !strings.HasPrefix(requestBody, strings.Repeat("x", maxBodyBytes)) || !strings.Contains(requestBody, "100 more bytes") {
		t.Fatalf("body was not capped at %d bytes (length %d)", maxBodyBytes, len(requestBody))
	}
}

func TestDetailDiagnosticsAreBounded(t *testing.T) {
	headers := make(http.Header)
	for i := 0; i < maxHeaderEntries+5; i++ {
		headers.Set(fmt.Sprintf("X-%03d", i), strings.Repeat("h", maxHeaderBytes+10))
	}
	logs := make([]store.LogEntry, maxDetailEntries+5)
	queries := make([]store.SQLQuery, maxDetailEntries+5)
	calls := make([]store.HTTPCall, maxDetailEntries+5)
	bottlenecks := make([]store.Bottleneck, maxDetailEntries+5)
	for i := range logs {
		logs[i] = store.LogEntry{Message: strings.Repeat("l", maxHeaderBytes+10)}
		queries[i] = store.SQLQuery{Query: strings.Repeat("q", maxHeaderBytes+10)}
		calls[i] = store.HTTPCall{URL: strings.Repeat("u", maxHeaderBytes+10)}
		bottlenecks[i] = store.Bottleneck{Description: strings.Repeat("b", maxHeaderBytes+10)}
	}
	d := detail(&store.RequestLog{
		RequestHeaders: headers,
		Logs:           logs,
		PanicStack:     strings.Repeat("p", maxTextBytes+10),
		PerformanceMetrics: &store.PerformanceMetrics{
			SQLQueries:  queries,
			HTTPCalls:   calls,
			Bottlenecks: bottlenecks,
		},
	}, maxBodyBytes*10)
	if len(d.RequestHeaders) != maxHeaderEntries || len(d.Logs) != maxDetailEntries ||
		len(d.SQLQueries) != maxDetailEntries || len(d.HTTPCalls) != maxDetailEntries ||
		len(d.Bottlenecks) != maxDetailEntries {
		t.Fatalf("diagnostic limits not applied: headers=%d logs=%d sql=%d http=%d bottlenecks=%d",
			len(d.RequestHeaders), len(d.Logs), len(d.SQLQueries), len(d.HTTPCalls), len(d.Bottlenecks))
	}
	if !strings.Contains(d.PanicStack, "more bytes") || !strings.Contains(d.SQLQueries[0].Query, "more bytes") {
		t.Fatalf("truncation was not disclosed: panic=%d query=%q", len(d.PanicStack), d.SQLQueries[0].Query)
	}
}

func TestDetailBoundsHeaderNames(t *testing.T) {
	longName := strings.Repeat("X", maxHeaderBytes+10)
	headers := flattenHeaders(http.Header{longName: []string{"value"}})
	for key := range headers {
		if len(key) <= maxHeaderBytes || !strings.Contains(key, "more bytes") {
			t.Fatalf("header name was not bounded with disclosure: length=%d key=%q", len(key), key)
		}
	}
}

func TestStatsRoutesAreCapped(t *testing.T) {
	st := store.NewMemory(maxListLimit + 5)
	for i := 0; i < maxListLimit+5; i++ {
		if err := st.Add(&store.RequestLog{ID: fmt.Sprintf("id-%d", i), Method: http.MethodGet, Path: fmt.Sprintf("/route/%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	out := call(t, connect(t, st, nil), "get_stats", struct{}{})
	if routes := out["routes"].([]any); len(routes) != maxListLimit {
		t.Fatalf("got %d routes, want cap %d", len(routes), maxListLimit)
	}
	if out["total_routes"].(float64) != maxListLimit+5 {
		t.Fatalf("total_routes = %v", out["total_routes"])
	}
}

func TestDiffReplayComparesBeyondExcerpt(t *testing.T) {
	body := strings.Repeat("x", defaultBodyBytes+1024)
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	defer app.Close()
	u, _ := url.Parse(app.URL)
	st := store.NewMemory(10)
	st.Add(&store.RequestLog{
		ID:           "long-response",
		Timestamp:    time.Now(),
		Method:       http.MethodGet,
		Path:         "/",
		StatusCode:   http.StatusOK,
		ResponseBody: body,
		Host:         u.Host,
	})
	s := connect(t, st, &config{baseURL: app.URL})
	out := call(t, s, "diff_replay", map[string]any{"id": "long-response"})
	if changed, _ := out["body_changed"].(bool); changed {
		t.Fatalf("identical body beyond 2 KiB excerpt reported changed: %+v", out)
	}
}

func TestDiffReplayReportsTruncatedAndUnavailableComparisons(t *testing.T) {
	prefix := "captured-prefix"
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/empty-capture" {
			w.Write([]byte("new body"))
			return
		}
		w.Write([]byte(prefix + " plus uncaptured bytes"))
	}))
	defer app.Close()

	st := store.NewMemory(2)
	for _, entry := range []*store.RequestLog{
		{ID: "truncated", Method: http.MethodGet, Path: "/truncated", StatusCode: http.StatusOK, ResponseBody: prefix + captureTruncationMarker},
		{ID: "empty", Method: http.MethodGet, Path: "/empty-capture", StatusCode: http.StatusOK},
	} {
		if err := st.Add(entry); err != nil {
			t.Fatal(err)
		}
	}
	s := connect(t, st, &config{baseURL: app.URL})

	truncated := call(t, s, "diff_replay", map[string]any{"id": "truncated"})
	if truncated["body_changed"] != false || truncated["body_compared"] != true || truncated["body_comparison_truncated"] != true {
		t.Fatalf("truncated comparison = %+v", truncated)
	}
	if !strings.Contains(truncated["summary"].(string), "truncated") {
		t.Fatalf("truncated summary = %q", truncated["summary"])
	}

	unavailable := call(t, s, "diff_replay", map[string]any{"id": "empty"})
	if unavailable["body_changed"] != false || unavailable["body_compared"] != false {
		t.Fatalf("unavailable comparison = %+v", unavailable)
	}
	if !strings.Contains(unavailable["summary"].(string), "unavailable") {
		t.Fatalf("unavailable summary = %q", unavailable["summary"])
	}
}

func TestClearRequests(t *testing.T) {
	st := seedStore(t)
	s := connect(t, st, nil)
	out := call(t, s, "clear_requests", struct{}{})
	if out["cleared"] != true {
		t.Fatalf("expected cleared, got %+v", out)
	}
	if len(st.GetAll()) != 0 {
		t.Fatal("store not cleared")
	}
}
