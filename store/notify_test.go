package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

type committedErrorStore struct {
	store.Store
}

func (s committedErrorStore) Add(log *store.RequestLog) error {
	if err := s.Store.Add(log); err != nil {
		return err
	}
	return errors.New("cleanup failed after insert")
}

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

func TestNotifyingStoreSignalsWhenAddReturnsAfterPersisting(t *testing.T) {
	wrapped := committedErrorStore{Store: store.NewMemory(10)}
	ns := store.WithNotify(wrapped)
	ch, cancel := ns.Subscribe()
	defer cancel()

	err := ns.Add(&store.RequestLog{ID: "a", Timestamp: time.Now()})
	if err == nil || err.Error() != "cleanup failed after insert" {
		t.Fatalf("Add error = %v", err)
	}
	if _, ok := ns.Get("a"); !ok {
		t.Fatal("expected the wrapped Add to persist before returning its error")
	}
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected a signal when Add persisted before returning an error")
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
