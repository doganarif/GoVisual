package store

import (
	"os"
	"testing"
)

func TestMongoStoage(t *testing.T) {
	os.Setenv("MONGO_URI", "mongodb://root:root@localhost:27017")
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI not set; skipping MongodB test")
	}
	store, err := NewMongoDBStore(uri, "logs", "request_logs", 10)
	if err != nil {
		t.Fatalf("failed to create MongoDB store: %v", err)
	}
	runStoreTests(t, store)
}
