package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestHandlerGates(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	h := Handler(store.NewMemory(10))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.9:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("remote request: got %d, want 403", rec.Code)
	}

	h = Handler(store.NewMemory(10), WithToken("s3cret"))
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: got %d, want 401", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer s3cret")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorized initialize: got %d body=%s", rec.Code, rec.Body.String())
	}
}
