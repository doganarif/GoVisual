package govisual

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlogHandlerAttachesLogsToRequest(t *testing.T) {
	logger := slog.New(SlogHandler(slog.NewTextHandler(io.Discard, nil)))

	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "cache miss", "key", "user:42")
	})
	h := Wrap(mux)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/work", nil))

	body := capturedRequests(t, h)
	if !strings.Contains(body, "cache miss") || !strings.Contains(body, "user:42") {
		t.Fatalf("captured request is missing the log line: %s", body)
	}
}

func TestSlogHandlerWithoutRequestContextPassesThrough(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(SlogHandler(slog.NewTextHandler(&buf, nil)))

	logger.Info("startup", "port", 8080)

	if !strings.Contains(buf.String(), "startup") {
		t.Fatalf("base handler did not receive the record: %s", buf.String())
	}
}

func TestSlogHandlerGroupsFlattenInCapture(t *testing.T) {
	logger := slog.New(SlogHandler(slog.NewTextHandler(io.Discard, nil))).
		WithGroup("db").With("driver", "postgres")

	mux := http.NewServeMux()
	mux.HandleFunc("/q", func(w http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "query ran", "rows", 3)
	})
	h := Wrap(mux)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/q", nil))

	body := capturedRequests(t, h)
	if !strings.Contains(body, "db.driver") || !strings.Contains(body, "db.rows") {
		t.Fatalf("expected flattened group keys in capture: %s", body)
	}
}
