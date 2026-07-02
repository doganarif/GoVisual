package store_test

import (
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

func TestNotifyingStoreSignalsOnAdd(t *testing.T) {
	ns := store.WithNotify(store.NewMemory(10))
	ch, cancel := ns.Subscribe()
	defer cancel()

	ns.Add(&store.RequestLog{ID: "a", Timestamp: time.Now()})

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected a signal after Add")
	}

	if _, ok := ns.Get("a"); !ok {
		t.Fatal("expected Add to reach the wrapped store")
	}
}

func TestNotifyingStoreCoalescesBursts(t *testing.T) {
	ns := store.WithNotify(store.NewMemory(10))
	ch, cancel := ns.Subscribe()
	defer cancel()

	for i := 0; i < 5; i++ {
		ns.Add(&store.RequestLog{ID: string(rune('a' + i)), Timestamp: time.Now()})
	}

	// At least one signal must be pending; draining should not block.
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected a coalesced signal")
	}
}

func TestNotifyingStoreCancelStopsDelivery(t *testing.T) {
	ns := store.WithNotify(store.NewMemory(10))
	ch, cancel := ns.Subscribe()
	cancel()

	ns.Add(&store.RequestLog{ID: "a", Timestamp: time.Now()})

	select {
	case <-ch:
		t.Fatal("cancelled subscriber should not receive signals")
	case <-time.After(50 * time.Millisecond):
	}
}
