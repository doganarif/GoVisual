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
}
