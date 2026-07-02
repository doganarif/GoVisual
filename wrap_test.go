package govisual

import (
	"net/http"
	"net/http/httptest"
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
