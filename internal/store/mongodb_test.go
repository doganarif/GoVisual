package store

import (
	"os"
	"testing"
)

func TestMongoStoage(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI not set; skipping MongoDB test")
	}
	store, err := NewMongoDBStore(uri, "logs", "request_logs", 10)
	if err != nil {
		t.Fatalf("failed to create MongoDB store: %v", err)
	}
	runStoreTests(t, store)
}
