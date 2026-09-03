package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
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

func TestAddReturnsSerializationErrors(t *testing.T) {
	s := &Store{}
	err := s.Add(&store.RequestLog{
		MiddlewareTrace: []map[string]interface{}{{"unsupported": make(chan int)}},
	})
	if err == nil || !strings.Contains(err.Error(), "marshal middleware trace") {
		t.Fatalf("Add error = %v, want middleware serialization error", err)
	}
}

func TestPostgresStoreMigratesHostColumn(t *testing.T) {
	connStr := os.Getenv("PG_CONN")
	if connStr == "" {
		t.Skip("PG_CONN not set; skipping PostgreSQL migration test")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	defer db.Close()

	tableName := fmt.Sprintf("logs_host_migration_%d", time.Now().UnixNano())
	_, err = db.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE,
			method TEXT,
			path TEXT,
			query TEXT,
			request_headers JSONB,
			response_headers JSONB,
			status_code INTEGER,
			duration BIGINT,
			request_body TEXT,
			response_body TEXT,
			error TEXT,
			middleware_trace JSONB,
			route_trace JSONB,
			extras JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`, tableName))
	if err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	defer func() {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE %s", tableName)); err != nil {
			t.Errorf("drop migration test table: %v", err)
		}
	}()

	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s (
			id, timestamp, method, path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error,
			middleware_trace, route_trace, extras
		) VALUES ('legacy-1', NOW(), 'GET', '/legacy', '', '{}', '{}', 200, 0, '', '', '', '[]', '{}', '{}')
	`, tableName))
	if err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	s, err := New(connStr, tableName, 10)
	if err != nil {
		t.Fatalf("open legacy table: %v", err)
	}
	defer s.Close()

	legacy, ok := s.Get("legacy-1")
	if !ok || legacy.Host != "" {
		t.Fatalf("legacy row after migration: %+v", legacy)
	}

	if err := s.Add(&store.RequestLog{
		ID:         "new-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Host:       "migrated.example.test",
		Path:       "/new",
		RawPath:    "/%6eew",
		StatusCode: 200,
	}); err != nil {
		t.Fatalf("add row after migration: %v", err)
	}
	got, ok := s.Get("new-1")
	if !ok || got.Host != "migrated.example.test" || got.RawPath != "/%6eew" {
		t.Fatalf("host dropped after migration: %+v", got)
	}
}
