package store

import (
	"fmt"
	"time"

	"database/sql"
	"testing"

	"github.com/doganarif/govisual/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteStore(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	defer db.Close()

	store, err := NewSQLiteStoreWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	runStoreTests(t, store)
}

func TestSQLiteStoreCleanupKeepsNewest(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	defer db.Close()

	store, err := NewSQLiteStoreWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	base := time.Now()
	for i := 0; i < 25; i++ {
		store.Add(&model.RequestLog{
			ID:        fmt.Sprintf("req-%02d", i),
			Timestamp: base.Add(time.Duration(i) * time.Second),
			Method:    "GET",
			Path:      "/x",
		})
	}

	store.cleanup()

	all := store.GetAll()
	if len(all) != 10 {
		t.Fatalf("expected capacity 10 after cleanup, got %d", len(all))
	}
	for _, l := range all {
		if l.ID < "req-15" {
			t.Fatalf("old entry %s survived cleanup", l.ID)
		}
	}
}
