package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
	"github.com/doganarif/govisual/v2/store/storetest"
)

func TestRedisStore(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	s, err := New(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}

	storetest.Run(t, s)
}

func TestRedisStoreOrderStableForEqualTimestamps(t *testing.T) {
	connStr := os.Getenv("REDIS_CONN")
	if connStr == "" {
		t.Skip("REDIS_CONN not set; skipping Redis test")
	}

	s, err := New(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer s.Close()
	defer s.Clear()

	// Same timestamp for every entry: order must come from the sorted set
	// (reverse-lexicographic on ID for equal scores), not map iteration.
	ts := time.Now()
	for _, id := range []string{"a", "b", "c", "d"} {
		s.Add(&store.RequestLog{ID: id, Timestamp: ts, Method: "GET", Path: "/x"})
	}

	want := []string{"d", "c", "b", "a"}
	for run := 0; run < 5; run++ {
		all := s.GetAll()
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

	s, err := New(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer s.Close()
	defer s.Clear()

	s.Add(&store.RequestLog{ID: "keep", Timestamp: time.Now(), Method: "GET", Path: "/x"})
	s.Add(&store.RequestLog{ID: "gone", Timestamp: time.Now(), Method: "GET", Path: "/x"})

	// Simulate TTL expiry: the key vanishes but the sorted set still has the ID.
	ctx := context.Background()
	if err := s.client.Del(ctx, s.keyPrefix+"gone").Err(); err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	all := s.GetAll()
	if len(all) != 1 || all[0].ID != "keep" {
		t.Fatalf("expected only 'keep', got %v", ids(all))
	}

	card, err := s.client.ZCard(ctx, s.keyPrefix+"logs").Result()
	if err != nil {
		t.Fatalf("zcard: %v", err)
	}
	if card != 1 {
		t.Fatalf("expected expired ID pruned from sorted set, ZCARD = %d", card)
	}
}

func ids(logs []*store.RequestLog) []string {
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

	s, err := New(connStr, 10, 3600)
	if err != nil {
		t.Fatalf("failed to create Redis store: %v", err)
	}
	defer s.Close()
	defer s.Clear()

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
