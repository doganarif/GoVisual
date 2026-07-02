package sqlite

import (
	"fmt"
	"time"

	"database/sql"
	"testing"

	"github.com/doganarif/govisual/v2/store"
	"github.com/doganarif/govisual/v2/store/storetest"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteStore(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	defer db.Close()

	s, err := NewWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	storetest.Run(t, s)
}

func TestSQLiteStoreCleanupKeepsNewest(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	defer db.Close()

	s, err := NewWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	base := time.Now()
	for i := 0; i < 25; i++ {
		s.Add(&store.RequestLog{
			ID:        fmt.Sprintf("req-%02d", i),
			Timestamp: base.Add(time.Duration(i) * time.Second),
			Method:    "GET",
			Path:      "/x",
		})
	}

	s.cleanup()

	all := s.GetAll()
	if len(all) != 10 {
		t.Fatalf("expected capacity 10 after cleanup, got %d", len(all))
	}
	for _, l := range all {
		if l.ID < "req-15" {
			t.Fatalf("old entry %s survived cleanup", l.ID)
		}
	}
}
