package sqlite

import (
	"database/sql"
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
	// Pre-v2 schema — no `extras` column.
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
	// A v2 row should now round trip.
	s.Add(&store.RequestLog{
		ID:         "new-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/v2",
		StatusCode: 500,
		PanicStack: "goroutine 1 [running]",
		Logs:       []store.LogEntry{{Level: "ERROR", Message: "hi"}},
	})
	back, ok := s.Get("new-1")
	if !ok || back.PanicStack == "" || len(back.Logs) != 1 {
		t.Fatalf("v2 fields dropped on upgraded table: %+v", back)
	}
}
