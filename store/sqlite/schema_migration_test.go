package sqlite

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
	_ "github.com/mattn/go-sqlite3"
)

func TestOpensAndUpgradesOldSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Pre-v2 schema with no `extras`, `host`, or `raw_path` column.
	_, err = db.Exec(`CREATE TABLE logs (
		id TEXT PRIMARY KEY,
		timestamp DATETIME,
		method TEXT,
		path TEXT,
		query TEXT,
		request_headers TEXT,
		response_headers TEXT,
		status_code INTEGER,
		duration INTEGER,
		request_body TEXT,
		response_body TEXT,
		error TEXT,
		middleware_trace TEXT,
		route_trace TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	// Insert something with pre-v2 fields only.
	_, err = db.Exec(`INSERT INTO logs (id, timestamp, method, path, query, request_headers, response_headers, status_code, duration, request_body, response_body, error, middleware_trace, route_trace)
		VALUES ('old-1', ?, 'GET', '/pre-v2', '', '{}', '{}', 200, 0, '', '', '', '[]', '{}')`, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("open against old schema: %v", err)
	}
	got, ok := s.Get("old-1")
	if !ok || got.Path != "/pre-v2" {
		t.Fatalf("pre-v2 row missing after upgrade: %+v", got)
	}
	if got.Host != "" {
		t.Fatalf("legacy row host = %q, want empty", got.Host)
	}

	for _, column := range []string{"extras", "host", "raw_path"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('logs') WHERE name = ?`, column).Scan(&count); err != nil {
			t.Fatalf("inspect %s migration: %v", column, err)
		}
		if count != 1 {
			t.Fatalf("expected one %s column after migration, got %d", column, count)
		}
	}

	// A v2 row should now round trip.
	if err := s.Add(&store.RequestLog{
		ID:         "new-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Host:       "migrated.example.test:9443",
		Path:       "/v2",
		RawPath:    "/v%32",
		StatusCode: 500,
		PanicStack: "goroutine 1 [running]",
		Logs:       []store.LogEntry{{Level: "ERROR", Message: "hi"}},
	}); err != nil {
		t.Fatalf("add row after migration: %v", err)
	}
	back, ok := s.Get("new-1")
	if !ok || back.Host != "migrated.example.test:9443" || back.RawPath != "/v%32" || back.PanicStack == "" || len(back.Logs) != 1 {
		t.Fatalf("v2 fields dropped on upgraded table: %+v", back)
	}

	// Cleanup must remain compatible with old tables that never had the
	// created_at column used by early v2 schemas.
	for i := 0; i < cleanupEveryN-1; i++ {
		if err := s.Add(&store.RequestLog{
			ID:        fmt.Sprintf("cleanup-%02d", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
			Method:    "GET",
			Path:      "/cleanup",
		}); err != nil {
			t.Fatalf("add %d to upgraded schema: %v", i, err)
		}
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM logs`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count > 10 {
		t.Fatalf("cleanup retained %d rows, want at most 10", count)
	}
}
