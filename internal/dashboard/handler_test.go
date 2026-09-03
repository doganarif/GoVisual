package dashboard

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

func TestSSEPushesOnAdd(t *testing.T) {
	ns := store.WithNotify(store.NewMemory(10))
	srv := httptest.NewServer(NewHandler(ns, nil, HandlerOptions{}))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatalf("connect SSE: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("SSE exposed wildcard CORS header %q", got)
	}

	events := make(chan string, 16)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") {
				events <- strings.TrimPrefix(line, "event: ")
			}
		}
	}()

	select {
	case ev := <-events:
		if ev != "snapshot" {
			t.Fatalf("first event = %q, want snapshot", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no snapshot event")
	}

	ns.Add(&store.RequestLog{ID: "x1", Timestamp: time.Now(), Method: "GET", Path: "/p"})

	// The push must arrive well under the 15s heartbeat tick.
	select {
	case ev := <-events:
		if ev != "append" {
			t.Fatalf("second event = %q, want append", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("append was not pushed; SSE still waiting on the ticker")
	}
}

func TestReplayUsesCapturedRequestAndPinnedDashboardOrigin(t *testing.T) {
	st := store.NewMemory(10)
	st.Add(&store.RequestLog{
		ID:          "captured-1",
		Method:      http.MethodPost,
		Path:        "/original",
		RawPath:     "/orig%69nal",
		Query:       "from=capture",
		RequestBody: "captured body",
		RequestHeaders: http.Header{
			"X-Keep":          []string{"captured"},
			"X-Remove":        []string{"captured"},
			"Accept-Encoding": []string{"br"},
			"Authorization":   []string{"[redacted by govisual]"},
			"Connection":      []string{"X-Captured-Hop"},
			"X-Captured-Hop":  []string{"remove me"},
			"Host":            []string{"attacker.invalid"},
			"Content-Length":  []string{"999"},
		},
	})

	type received struct {
		method        string
		path          string
		escapedPath   string
		query         string
		body          string
		header        http.Header
		host          string
		contentLength int64
	}
	got := make(chan received, 3)
	dashboardHandler := NewHandler(st, nil, HandlerOptions{EnableReplay: true, AllowLocalhostReplay: true})
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/dashboard/") {
			http.StripPrefix("/dashboard", dashboardHandler).ServeHTTP(w, r)
			return
		}
		body, _ := io.ReadAll(r.Body)
		got <- received{method: r.Method, path: r.URL.Path, escapedPath: r.URL.EscapedPath(), query: r.URL.RawQuery, body: string(body), header: r.Header.Clone(), host: r.Host, contentLength: r.ContentLength}
		w.Header().Set("X-Replayed", "yes")
		w.Header().Set("Set-Cookie", "session=secret")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("replayed"))
	})
	srv := httptest.NewServer(root)
	defer srv.Close()

	postReplay := func(payload map[string]any) received {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(srv.URL+"/dashboard/api/replay", "application/json", strings.NewReader(string(body)))
		if err != nil {
			t.Fatalf("replay: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			t.Fatalf("replay returned %d: %s", resp.StatusCode, data)
		}
		var result struct {
			StatusCode int         `json:"statusCode"`
			Headers    http.Header `json:"headers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("decode replay response: %v", err)
		}
		if result.StatusCode != http.StatusAccepted {
			t.Fatalf("replayed status = %d, want %d", result.StatusCode, http.StatusAccepted)
		}
		if got := result.Headers.Get("Set-Cookie"); got != "[redacted by govisual]" {
			t.Fatalf("replay response Set-Cookie = %q, want redaction", got)
		}
		select {
		case req := <-got:
			return req
		case <-time.After(time.Second):
			t.Fatal("pinned application did not receive replay")
			return received{}
		}
	}

	first := postReplay(map[string]any{
		"requestId": "captured-1",
		"url":       "http://attacker.invalid/stolen?legacy=1",
	})
	if first.method != http.MethodPost || first.path != "/stolen" || first.query != "legacy=1" || first.body != "captured body" {
		t.Fatalf("legacy URL shape not applied to pinned destination: %+v", first)
	}
	if first.host != strings.TrimPrefix(srv.URL, "http://") {
		t.Fatalf("replay Host = %q, want dashboard host", first.host)
	}
	if first.header.Get("Authorization") != "" {
		t.Fatalf("stored redacted authorization was forwarded: %v", first.header)
	}
	if first.header.Get("Accept-Encoding") != "br" {
		t.Fatalf("end-to-end Accept-Encoding was not preserved: %v", first.header)
	}

	second := postReplay(map[string]any{
		"requestId": "captured-1",
		"method":    http.MethodPut,
		"path":      "/changed?from=override",
		"body":      "",
		"headers": map[string]string{
			"X-Keep":         "override",
			"Connection":     "X-Override-Hop",
			"X-Override-Hop": "remove me too",
			"Host":           "attacker.invalid",
			"Content-Length": "1234",
			"Authorization":  "Bearer deliberate-override",
			"X-Api-Key":      "[redacted by govisual]",
		},
	})
	if second.method != http.MethodPut || second.path != "/changed" || second.query != "from=override" || second.body != "" {
		t.Fatalf("replay overrides not applied: %+v", second)
	}
	if second.header.Get("X-Keep") != "override" || second.header.Get("X-Govisual-Replay") != "1" {
		t.Fatalf("safe headers missing: %v", second.header)
	}
	for _, name := range []string{"Connection", "X-Captured-Hop", "X-Override-Hop"} {
		if value := second.header.Get(name); value != "" {
			t.Fatalf("hop-by-hop header %s forwarded as %q", name, value)
		}
	}
	if second.contentLength != 0 {
		t.Fatalf("caller-controlled Content-Length was forwarded: got %d", second.contentLength)
	}
	if second.host != strings.TrimPrefix(srv.URL, "http://") {
		t.Fatalf("Host override escaped pinned destination: %q", second.host)
	}
	if second.header.Get("Authorization") != "Bearer deliberate-override" {
		t.Fatalf("explicit authorization override missing: %v", second.header)
	}
	if second.header.Get("X-Api-Key") != "" {
		t.Fatalf("redaction placeholder override was forwarded: %v", second.header)
	}
	if second.header.Get("X-Remove") != "" {
		t.Fatalf("header removed in the dashboard was still forwarded: %v", second.header)
	}

	third := postReplay(map[string]any{"requestId": "captured-1"})
	if third.path != "/original" || third.escapedPath != "/orig%69nal" {
		t.Fatalf("captured encoded path was not preserved: %+v", third)
	}
}

func TestReplayRequiresExistingRequestAndRelativePath(t *testing.T) {
	st := store.NewMemory(10)
	st.Add(&store.RequestLog{ID: "one", Method: http.MethodGet, Path: "/ok"})
	st.Add(&store.RequestLog{ID: "truncated", Method: http.MethodPost, Path: "/ok", RequestBody: "prefix" + captureTruncationMarker})
	h := NewHandler(st, nil, HandlerOptions{EnableReplay: true, ReplayBaseURL: "http://localhost:1"})

	for _, tc := range []struct {
		name    string
		payload string
		want    int
	}{
		{name: "missing id", payload: `{}`, want: http.StatusBadRequest},
		{name: "unknown id", payload: `{"requestId":"missing"}`, want: http.StatusNotFound},
		{name: "absolute url path", payload: `{"requestId":"one","path":"http://attacker.invalid/x"}`, want: http.StatusBadRequest},
		{name: "authority path", payload: `{"requestId":"one","path":"//attacker.invalid/x"}`, want: http.StatusBadRequest},
		{name: "truncated body", payload: `{"requestId":"truncated"}`, want: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/replay", strings.NewReader(tc.payload))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d, want %d: %s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}
