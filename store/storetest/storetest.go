// Package storetest exercises a store.Store implementation against the
// behavior every backend must provide. Storage modules use it in their own
// test suites.
package storetest

import (
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

// Run puts a store through the shared conformance checks. It clears and
// closes the store when done.
func Run(t *testing.T, s store.Store) {
	t.Helper()
	defer s.Close()
	defer s.Clear()

	log := &store.RequestLog{
		ID:         "test-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
	}

	s.Add(log)

	got, ok := s.Get("test-1")
	if !ok || got.ID != "test-1" {
		t.Errorf("expected to get log with ID 'test-1', got %+v", got)
	}

	all := s.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 log in GetAll, got %d", len(all))
	}

	latest := s.GetLatest(1)
	if len(latest) != 1 || latest[0].ID != "test-1" {
		t.Errorf("expected to get latest log with ID 'test-1', got %+v", latest)
	}

	if err := s.Clear(); err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	if len(s.GetAll()) != 0 {
		t.Error("expected no logs after Clear")
	}

	// Every v2 capture field must survive a round trip. Backends added their
	// own columns for these historically; forgetting to persist Logs or
	// PanicStack is a real regression, hence the explicit contract.
	rich := &store.RequestLog{
		ID:         "rich-1",
		Timestamp:  time.Now(),
		Method:     "POST",
		Path:       "/api/things",
		StatusCode: 500,
		Error:      "panic: kaboom",
		PanicStack: "goroutine 1 [running]:\nmain.boom()\n\t/tmp/x.go:5",
		Logs: []store.LogEntry{
			{Time: time.Now(), Level: "ERROR", Message: "before panic", Attrs: map[string]any{"k": "v"}},
		},
		PerformanceMetrics: &store.PerformanceMetrics{
			RequestID:  "rich-1",
			Duration:   42 * time.Millisecond,
			SQLQueries: []store.SQLQuery{{Query: "SELECT 1", Duration: time.Millisecond, Rows: 1}},
			HTTPCalls:  []store.HTTPCall{{Method: "GET", URL: "http://x", Duration: time.Millisecond, Status: 200}},
		},
	}
	if err := s.Add(rich); err != nil {
		t.Fatalf("Add rich: %v", err)
	}
	back, ok := s.Get("rich-1")
	if !ok {
		t.Fatal("expected rich-1 after Add")
	}
	if back.PanicStack == "" {
		t.Error("PanicStack dropped by store round trip")
	}
	if len(back.Logs) != 1 || back.Logs[0].Message != "before panic" {
		t.Errorf("Logs dropped by store round trip: %+v", back.Logs)
	}
	if back.PerformanceMetrics == nil ||
		len(back.PerformanceMetrics.SQLQueries) != 1 ||
		len(back.PerformanceMetrics.HTTPCalls) != 1 {
		t.Errorf("PerformanceMetrics dropped or trimmed: %+v", back.PerformanceMetrics)
	}
	if err := s.Clear(); err != nil {
		t.Errorf("Clear after rich: %v", err)
	}
}
