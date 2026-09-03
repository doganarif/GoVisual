package govisual

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

type addErrorStore struct {
	store.Store
	err error
}

func (s *addErrorStore) Add(*store.RequestLog) error { return s.err }

func dashboardStatus(t *testing.T, remoteAddr string, opts ...Option) int {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {})
	h := Wrap(mux, opts...)

	req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
	req.RemoteAddr = remoteAddr
	req.Host = "localhost"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestDashboardLocalhostOnlyByDefault(t *testing.T) {
	if got := dashboardStatus(t, "127.0.0.1:5555"); got != http.StatusOK {
		t.Fatalf("loopback request: got %d, want 200", got)
	}
	if got := dashboardStatus(t, "203.0.113.7:5555"); got != http.StatusForbidden {
		t.Fatalf("remote request: got %d, want 403", got)
	}
}

func TestDashboardAllowRemote(t *testing.T) {
	if got := dashboardStatus(t, "203.0.113.7:5555", WithAllowRemote()); got != http.StatusOK {
		t.Fatalf("remote request with WithAllowRemote: got %d, want 200", got)
	}
}

func TestDashboardRejectsUntrustedHostAndCrossOrigin(t *testing.T) {
	mux := http.NewServeMux()
	h := Wrap(mux)

	tests := []struct {
		name   string
		host   string
		origin string
		want   int
	}{
		{name: "localhost", host: "localhost:8080", want: http.StatusOK},
		{name: "loopback ipv4", host: "127.0.0.1:8080", origin: "http://127.0.0.1:8080", want: http.StatusOK},
		{name: "loopback ipv6", host: "[::1]:8080", origin: "http://[::1]:8080", want: http.StatusOK},
		{name: "localhost subdomain", host: "app.localhost:8080", origin: "http://app.localhost:8080", want: http.StatusOK},
		{name: "dns rebinding host", host: "attacker.example", want: http.StatusForbidden},
		{name: "localhost suffix trick", host: "localhost.attacker.example", want: http.StatusForbidden},
		{name: "cross origin", host: "localhost:8080", origin: "http://attacker.example", want: http.StatusForbidden},
		{name: "different port", host: "localhost:8080", origin: "http://localhost:8081", want: http.StatusForbidden},
		{name: "opaque origin", host: "localhost:8080", origin: "null", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
			req.RemoteAddr = "127.0.0.1:5555"
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("got %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestDashboardRedirectIsGuarded(t *testing.T) {
	h := Wrap(http.NewServeMux())
	req := httptest.NewRequest(http.MethodGet, "/__viz", nil)
	req.RemoteAddr = "127.0.0.1:5555"
	req.Host = "attacker.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unguarded dashboard redirect returned %d, want 403", rec.Code)
	}
}

func TestDashboardAllowRemoteStillRequiresSameOrigin(t *testing.T) {
	h := Wrap(http.NewServeMux(), WithAllowRemote())
	for _, tc := range []struct {
		origin string
		want   int
	}{
		{origin: "https://debug.example", want: http.StatusOK},
		{origin: "http://debug.example", want: http.StatusOK},
		{origin: "https://debug.example:444", want: http.StatusForbidden},
		{origin: "https://attacker.example", want: http.StatusForbidden},
	} {
		req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
		req.RemoteAddr = "203.0.113.7:5555"
		req.Host = "debug.example"
		req.Header.Set("Origin", tc.origin)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("origin %q: got %d, want %d", tc.origin, rec.Code, tc.want)
		}
	}
}

func capturedRequests(t *testing.T, h http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
	req.RemoteAddr = "127.0.0.1:5555"
	req.Host = "localhost"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Body.String()
}

func TestSampleRateZeroCapturesNothing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {})
	h := Wrap(mux, WithSampleRate(0))

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("sampled-out request must still be served, got %d", rec.Code)
		}
	}

	if body := capturedRequests(t, h); strings.Contains(body, "/hello") {
		t.Fatalf("sample rate 0 captured requests: %s", body)
	}
}

func TestSampleRateDefaultCapturesEverything(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {})
	h := Wrap(mux)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/hello", nil))

	if body := capturedRequests(t, h); !strings.Contains(body, "/hello") {
		t.Fatalf("default sampling missed a request: %s", body)
	}
}

func TestPanicIsCapturedAndRepropagated(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	})
	h := Wrap(mux)

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic must propagate through govisual")
			}
		}()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))
	}()

	body := capturedRequests(t, h)
	if !strings.Contains(body, "panic: kaboom") {
		t.Fatalf("panic not recorded: %s", body)
	}
	if !strings.Contains(body, "PanicStack") || !strings.Contains(body, "boom") {
		t.Fatalf("stack not recorded: %s", body)
	}
	if !strings.Contains(body, `"StatusCode":500`) {
		t.Fatalf("expected 500 on panicked request: %s", body)
	}
}

func TestPanicIsCapturedWithProfiling(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	})
	h := Wrap(mux, WithProfiling(true))

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic must propagate through govisual")
			}
		}()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))
	}()

	body := capturedRequests(t, h)
	if !strings.Contains(body, "panic: kaboom") || !strings.Contains(body, `"StatusCode":500`) {
		t.Fatalf("panic not recorded under profiling: %s", body)
	}
}

func TestEventAttachesToRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		Event(r.Context(), "cache miss", "key", "user:7")
	})
	h := Wrap(mux)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/work", nil))

	body := capturedRequests(t, h)
	if !strings.Contains(body, "cache miss") || !strings.Contains(body, "user:7") {
		t.Fatalf("event not attached: %s", body)
	}
}

func TestEventOutsideRequestIsNoOp(t *testing.T) {
	Event(t.Context(), "orphan event", "k", "v")
}

func TestFaviconIsIgnoredByDefault(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {})
	h := Wrap(mux)

	for _, path := range []string{"/hello", "/favicon.ico"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	body := capturedRequests(t, h)
	if !strings.Contains(body, "/hello") {
		t.Fatalf("expected /hello captured: %s", body)
	}
	if strings.Contains(body, "/favicon.ico") {
		t.Fatalf("expected /favicon.ico ignored: %s", body)
	}
}

func TestWithErrorHandlerReceivesStoreFailures(t *testing.T) {
	wantErr := errors.New("write failed")
	st := &addErrorStore{Store: store.NewMemory(1), err: wantErr}
	var gotErr error
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), WithStore(st), WithErrorHandler(func(err error) {
		gotErr = err
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/work", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("error handler got %v, want wrapped %v", gotErr, wantErr)
	}
}

func TestRemoteDashboardReplayRequiresExplicitBaseURL(t *testing.T) {
	st := store.NewMemory(2)
	if err := st.Add(&store.RequestLog{ID: "one", Method: http.MethodGet, Path: "/target"}); err != nil {
		t.Fatal(err)
	}
	h := Wrap(http.NewServeMux(), WithStore(st), WithAllowRemote(), WithReplayEnabled(true))
	req := httptest.NewRequest(http.MethodPost, "/__viz/api/replay", strings.NewReader(`{"requestId":"one"}`))
	req.RemoteAddr = "203.0.113.7:5555"
	req.Host = "debug.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "replay base URL is required") {
		t.Fatalf("remote replay without base = %d %q, want configuration rejection", rec.Code, rec.Body.String())
	}
}

func TestRemoteDashboardReplayUsesExplicitBaseURL(t *testing.T) {
	targeted := make(chan string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targeted <- r.URL.RequestURI()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	st := store.NewMemory(2)
	if err := st.Add(&store.RequestLog{ID: "one", Method: http.MethodGet, Path: "/target", Query: "x=1"}); err != nil {
		t.Fatal(err)
	}
	h := Wrap(http.NewServeMux(), WithStore(st), WithAllowRemote(), WithReplayEnabled(true), WithReplayBaseURL(target.URL))
	req := httptest.NewRequest(http.MethodPost, "/__viz/api/replay", strings.NewReader(`{"requestId":"one"}`))
	req.RemoteAddr = "203.0.113.7:5555"
	req.Host = "debug.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("remote replay with base = %d: %s", rec.Code, rec.Body.String())
	}
	select {
	case uri := <-targeted:
		if uri != "/target?x=1" {
			t.Fatalf("explicit replay target URI = %q", uri)
		}
	case <-time.After(time.Second):
		t.Fatal("configured replay target was not called")
	}
}
