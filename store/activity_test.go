package store_test

import (
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

func TestActivityLogRingEviction(t *testing.T) {
	log := store.NewActivityLog(3)
	for _, name := range []string{"a", "b", "c", "d"} {
		log.Record(store.ActivityEntry{Tool: name})
	}

	got := log.List(0)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	// Newest first.
	if got[0].Tool != "d" || got[1].Tool != "c" || got[2].Tool != "b" {
		t.Fatalf("order = [%s %s %s], want [d c b]", got[0].Tool, got[1].Tool, got[2].Tool)
	}
}

func TestActivityLogListLimit(t *testing.T) {
	log := store.NewActivityLog(10)
	for _, name := range []string{"a", "b", "c"} {
		log.Record(store.ActivityEntry{Tool: name})
	}
	got := log.List(2)
	if len(got) != 2 || got[0].Tool != "c" || got[1].Tool != "b" {
		t.Fatalf("limit=2 = %+v", got)
	}
}

func TestActivityLogSubscribe(t *testing.T) {
	log := store.NewActivityLog(5)
	ch, cancel := log.Subscribe()
	defer cancel()

	log.Record(store.ActivityEntry{Tool: "get_last_error"})

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected a signal from Subscribe after Record")
	}

	cancel()
	log.Record(store.ActivityEntry{Tool: "clear_requests", Mutating: true})
	select {
	case <-ch:
		t.Fatal("cancelled subscriber should not receive further signals")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestActivityLogTimeAutoFilled(t *testing.T) {
	log := store.NewActivityLog(5)
	log.Record(store.ActivityEntry{Tool: "x"})
	got := log.List(1)
	if got[0].Time.IsZero() {
		t.Fatal("Record should set Time when caller left it zero")
	}
}
