package store

import (
	"os"
	"testing"
)

func TestRedisStore(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	store, err := NewRedisStore(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}

	runStoreTests(t, store)
}
