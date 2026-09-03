package store

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequestLog(t *testing.T) {
	req, err := http.NewRequest("POST", "http://localhost:8080/test-path?foo=bar", strings.NewReader("body-content"))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Test-Header", "HeaderValue")

	log := NewRequestLog(req)

	if log.ID == "" {
		t.Error("expected ID to be generated, got empty string")
	}

	if log.Method != "POST" {
		t.Errorf("expected method to be POST, got %s", log.Method)
	}

	if log.Path != "/test-path" {
		t.Errorf("expected method to be /test-path, got %s", log.Path)
	}

	if log.Query != "foo=bar" {
		t.Errorf("expected query to be foo=bar, got %s", log.Query)
	}

	if log.RequestHeaders.Get("X-Test-Header") != "HeaderValue" {
		t.Errorf("expected request header to Header Value, got %s", log.RequestHeaders.Get("X-Test-Header"))
	}

	if log.Timestamp.IsZero() {
		t.Errorf("expected timestamp set, got zero value")
	}
}

func TestSetResponseHeadersCopiesEmptyMap(t *testing.T) {
	headers := make(http.Header)
	requestLog := &RequestLog{}
	requestLog.SetResponseHeaders(headers)

	headers.Set("Set-Cookie", "session=secret")
	if got := requestLog.ResponseHeaders.Get("Set-Cookie"); got != "" {
		t.Fatalf("response headers alias caller map and exposed %q", got)
	}
}

func TestNewRequestLogPreservesEscapedPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost/users/a%2Fb", nil)
	requestLog := NewRequestLog(req)
	if requestLog.Path != "/users/a/b" || requestLog.RawPath != "/users/a%2Fb" {
		t.Fatalf("captured path = %q raw=%q", requestLog.Path, requestLog.RawPath)
	}
}
