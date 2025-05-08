package model

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewRequestLog(t *testing.T) {
	req, err := http.NewRequest("POST", "http://localhost:8080/test-path?foo=bar", strings.NewReader("body-content"))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Test-Header", "HeaderValue")

	log := NewHTTPRequestLog(req)

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
