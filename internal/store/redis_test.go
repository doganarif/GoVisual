package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/doganarif/govisual/internal/model"
)

func TestRedisStore(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	store, err := NewRedisStore(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}

	runStoreTests(t, store)
}

func TestRedisStoreOrderStableForEqualTimestamps(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	store, err := NewRedisStore(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer store.Close()
	defer store.Clear()

	// Same timestamp for every entry: order must come from the sorted set
	// (reverse-lexicographic on ID for equal scores), not map iteration.
	ts := time.Now()
	for _, id := range []string{"a", "b", "c", "d"} {
		store.Add(&model.RequestLog{ID: id, Timestamp: ts, Method: "GET", Path: "/x"})
	}

	want := []string{"d", "c", "b", "a"}
	for run := 0; run < 5; run++ {
		all := store.GetAll()
		if len(all) != len(want) {
			t.Fatalf("expected %d logs, got %d", len(want), len(all))
		}
		for i, w := range want {
			if all[i].ID != w {
				t.Fatalf("run %d: order = %v, want %v", run, ids(all), want)
			}
		}
	}
}

func TestRedisStorePrunesExpiredIDs(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	store, err := NewRedisStore(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer store.Close()
	defer store.Clear()

	store.Add(&model.RequestLog{ID: "keep", Timestamp: time.Now(), Method: "GET", Path: "/x"})
	store.Add(&model.RequestLog{ID: "gone", Timestamp: time.Now(), Method: "GET", Path: "/x"})

	// Simulate TTL expiry: the key vanishes but the sorted set still has the ID.
	ctx := context.Background()
	if err := store.client.Del(ctx, store.keyPrefix+"gone").Err(); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	all := store.GetAll()
	if len(all) != 1 || all[0].ID != "keep" {
		t.Fatalf("expected only 'keep', got %v", ids(all))
	}

	card, err := store.client.ZCard(ctx, store.keyPrefix+"logs").Result()
	if err != nil {
		t.Fatalf("zcard: %v", err)
	}
	if card != 1 {
		t.Fatalf("expected expired ID pruned from sorted set, ZCARD = %d", card)
	}
}

func ids(logs []*model.RequestLog) []string {
	out := make([]string, len(logs))
	for i, l := range logs {
		out[i] = l.ID
	}
	return out
}

func TestRedisStoreCleanupKeepsNewest(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	store, err := NewRedisStore(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer store.Close()
	defer store.Clear()

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
