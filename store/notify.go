package store

import "sync"

// NotifyingStore wraps a Store so every successful Add signals subscribers.
// The dashboard uses it to push live updates instead of polling; anything
// that wants to react to new requests can Subscribe.
type NotifyingStore struct {
	Store
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

// WithNotify wraps s with change notification.
func WithNotify(s Store) *NotifyingStore {
	return &NotifyingStore{
		Store: s,
		subs:  make(map[chan struct{}]struct{}),
	}
}

func (n *NotifyingStore) Add(l *RequestLog) error {
	if err := n.Store.Add(l); err != nil {
		return err
	}
	n.mu.Lock()
	for ch := range n.subs {
		// Non-blocking: the buffered channel coalesces bursts, a slow
		// subscriber still sees one signal for the newest write.
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	n.mu.Unlock()
	return nil
}

// Subscribe returns a channel that receives a signal after each Add, and a
// cancel function that must be called when done. Signals carry no payload;
// read the store for the actual entries.
func (n *NotifyingStore) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	n.subs[ch] = struct{}{}
	n.mu.Unlock()

	cancel := func() {
		n.mu.Lock()
		delete(n.subs, ch)
		n.mu.Unlock()
	}
	return ch, cancel
}
