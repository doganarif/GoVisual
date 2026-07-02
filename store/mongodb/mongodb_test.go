package mongodb

import (
	"os"
	"testing"

	"github.com/doganarif/govisual/v2/store/storetest"
)

func TestMongoDBStore(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI not set; skipping MongoDB test")
	}

	s, err := New(uri, "logs", "request_logs", 10)
	if err != nil {
		t.Fatalf("failed to create MongoDB store: %v", err)
	}

	storetest.Run(t, s)
}
