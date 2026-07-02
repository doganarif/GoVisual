package govisual

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func dashboardStatus(t *testing.T, remoteAddr string, opts ...Option) int {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {})
	h := Wrap(mux, opts...)

	req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
	req.RemoteAddr = remoteAddr
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

func capturedRequests(t *testing.T, h http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
	req.RemoteAddr = "127.0.0.1:5555"
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
