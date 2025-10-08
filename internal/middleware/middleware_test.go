package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/options"
)

// mockStore implements store.Store for testing
type mockStore struct {
	logs []*model.RequestLog
}

func (m *mockStore) Add(log *model.RequestLog) {
	m.logs = append(m.logs, log)
}

func (m *mockStore) Get(id string) (*model.RequestLog, bool) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, true
		}
	}
	return nil, false
}

func (m *mockStore) GetAll() []*model.RequestLog {
	return m.logs
}

func (m *mockStore) Clear() error {
	m.logs = nil
	return nil
}

func (m *mockStore) GetLatest(n int) []*model.RequestLog {
	if n >= len(m.logs) {
		return m.logs
	}
	return m.logs[len(m.logs)-n:]
}

func (m *mockStore) Close() error {
	return nil
}

// mockPathMatcher implements PathMatcher
type mockPathMatcher struct{}

func (m *mockPathMatcher) ShouldIgnorePath(path string) bool {
	return false
}

func TestWrapMiddleware(t *testing.T) {
	store := &mockStore{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("hello world"))
	})

	config := &options.Config{
		LogRequestBody:  true,
		LogResponseBody: true,
	}

	wrapped := Wrap(handler, store, config, &mockPathMatcher{})

	req := httptest.NewRequest("POST", "/test?x=1", strings.NewReader("sample-body"))
	req.Header.Set("X-Test", "test")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if len(store.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(store.logs))
	}
	log := store.logs[0]

	if log.Method != "POST" {
		t.Errorf("expected Method POST, got %s", log.Method)
	}
	if log.Path != "/test" {
		t.Errorf("expected Path /test, got %s", log.Path)
	}
	if log.RequestBody != "sample-body" {
		t.Errorf("expected RequestBody to be 'sample-body', got '%s'", log.RequestBody)
	}
	if log.ResponseBody != "hello world" {
		t.Errorf("expected ResponseBody to be 'hello world', got '%s'", log.ResponseBody)
	}
	if log.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, log.StatusCode)
	}
	if log.Duration < 0 {
		t.Errorf("expected Duration > 0, got %d", log.Duration)
	}
}
