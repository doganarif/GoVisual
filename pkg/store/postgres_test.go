package store

import (
	"os"
	"testing"
)

func TestPostgresStore(t *testing.T) {
	connStr := os.Getenv("PG_CONN")
	if connStr == "" {
		t.Skip("PG_CONN not set; skipping PostgreSQL test")
	}

	store, err := NewPostgresStore(connStr, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create Postgres store: %v", err)
	}

	runStoreTests(t, store)
}
