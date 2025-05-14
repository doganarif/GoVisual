package store

import (
	"testing"
	"time"

	"github.com/doganarif/govisual/internal/model"
)

func runStoreTests(t *testing.T, store Store) {
	defer store.Clear()
	defer store.Close()

	// Create a log entry
	log := &model.RequestLog{
		ID:         "test-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
	}

	store.Add(log)

	// Test Get
	got, ok := store.Get("test-1")
	if !ok || got.ID != "test-1" {
		t.Errorf("expected to get log with ID 'test-1', got %+v", got)
	}

	// Test GetAll
	all := store.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 log in GetAll, got %d", len(all))
	}

	// Test GetLatest
	latest := store.GetLatest(1)
	if len(latest) != 1 || latest[0].ID != "test-1" {
		t.Errorf("expected to get latest log with ID 'test-1', got %+v", latest)
	}

	// Test Clear
	if err := store.Clear(); err != nil {
		t.Errorf("Clear failed: %v", err)
	}
	if len(store.GetAll()) != 0 {
		t.Error("expected no logs after Clear")
	}
}
