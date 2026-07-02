package postgres

import (
	"os"
	"testing"

	"github.com/doganarif/govisual/v2/store/storetest"
)

func TestPostgresStore(t *testing.T) {
	connStr := os.Getenv("PG_CONN")
	if connStr == "" {
		t.Skip("PG_CONN not set; skipping PostgreSQL test")
	}

	s, err := New(connStr, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create Postgres store: %v", err)
	}

	storetest.Run(t, s)
}
