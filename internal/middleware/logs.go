package middleware

import (
	"context"
	"sync"

	"github.com/doganarif/govisual/v2/store"
)

// maxLogsPerRequest bounds how many log lines a single request may retain,
// so a logging loop can't grow a request record without limit.
const maxLogsPerRequest = 200

type logCollectorKey struct{}

// LogCollector accumulates log lines emitted during one request.
type LogCollector struct {
	mu      sync.Mutex
	entries []store.LogEntry
	dropped int
}

// WithLogCollector attaches a fresh collector to ctx and returns both.
func WithLogCollector(ctx context.Context) (context.Context, *LogCollector) {
	c := &LogCollector{}
	return context.WithValue(ctx, logCollectorKey{}, c), c
}

// LogCollectorFromContext returns the request's collector, or nil when the
// context does not belong to a captured request.
func LogCollectorFromContext(ctx context.Context) *LogCollector {
	c, _ := ctx.Value(logCollectorKey{}).(*LogCollector)
	return c
}

// Append records one log line, dropping beyond the per-request cap.
func (c *LogCollector) Append(e store.LogEntry) {
	c.mu.Lock()
	if len(c.entries) < maxLogsPerRequest {
		c.entries = append(c.entries, e)
	} else {
		c.dropped++
	}
	c.mu.Unlock()
}

// Snapshot returns the collected lines. A trailing marker entry reports how
// many lines were dropped, if any.
func (c *LogCollector) Snapshot() []store.LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := append([]store.LogEntry(nil), c.entries...)
	if c.dropped > 0 {
		out = append(out, store.LogEntry{
			Level:   "WARN",
			Message: "govisual: log capture truncated",
			Attrs:   map[string]any{"dropped": c.dropped},
		})
	}
	return out
}
