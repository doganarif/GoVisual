package store

import (
	"sync"
	"time"
)

// ActivityEntry is one recorded action taken by a coding agent (or any MCP
// client) against the govisual store. Keeping this in the public store
// package lets Wrap and the mcp module share one bounded ring buffer.
type ActivityEntry struct {
	Time     time.Time         `json:"time"`
	Tool     string            `json:"tool"`
	Args     map[string]string `json:"args,omitempty"`
	Duration time.Duration     `json:"duration"`
	Error    string            `json:"error,omitempty"`
	// Mutating flags calls that changed dashboard state (replay, clear,
	// annotate). Read-only calls (list, get, search) are the default.
	Mutating bool `json:"mutating,omitempty"`
}

// ActivityLog is a bounded ring of recent agent tool calls. Safe for
// concurrent Record/List; entries returned by List are copies so callers
// can iterate without holding the lock.
type ActivityLog struct {
	mu       sync.Mutex
	capacity int
	entries  []ActivityEntry
	subs     map[chan struct{}]struct{}
}

// NewActivityLog creates a log capped at the given number of entries; new
// entries evict the oldest.
func NewActivityLog(capacity int) *ActivityLog {
	if capacity <= 0 {
		capacity = 200
	}
	return &ActivityLog{
		capacity: capacity,
		entries:  make([]ActivityEntry, 0, capacity),
		subs:     make(map[chan struct{}]struct{}),
	}
}

// Record appends an entry. Time defaults to now if the caller passes a zero
// value. Subscribers get a non-blocking signal.
func (a *ActivityLog) Record(e ActivityEntry) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	a.mu.Lock()
	if len(a.entries) == a.capacity {
		copy(a.entries, a.entries[1:])
		a.entries = a.entries[:len(a.entries)-1]
	}
	a.entries = append(a.entries, e)
	for ch := range a.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	a.mu.Unlock()
}

// List returns a copy of the entries, newest first, capped at limit
// (<= 0 returns all).
func (a *ActivityLog) List(limit int) []ActivityEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	n := len(a.entries)
	if limit > 0 && limit < n {
		n = limit
	}
	out := make([]ActivityEntry, n)
	for i := 0; i < n; i++ {
		out[i] = a.entries[len(a.entries)-1-i]
	}
	return out
}

// Subscribe returns a signal channel for Record events, plus a cancel
// function that must be called when done.
func (a *ActivityLog) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	a.mu.Lock()
	a.subs[ch] = struct{}{}
	a.mu.Unlock()
	return ch, func() {
		a.mu.Lock()
		delete(a.subs, ch)
		a.mu.Unlock()
	}
}
