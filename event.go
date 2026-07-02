package govisual

import (
	"context"
	"time"

	"github.com/doganarif/govisual/v2/internal/middleware"
	"github.com/doganarif/govisual/v2/store"
)

// Event annotates the current request with a named application event:
//
//	govisual.Event(r.Context(), "cache miss", "key", key, "tier", "redis")
//
// The event shows up in the request's log timeline, and in the middleware
// trace when profiling is enabled. Key/value pairs follow the slog
// convention; a call outside a captured request is a no-op.
func Event(ctx context.Context, name string, kv ...any) {
	var attrs map[string]any
	if len(kv) > 1 {
		attrs = make(map[string]any, len(kv)/2)
		for i := 0; i+1 < len(kv); i += 2 {
			if k, ok := kv[i].(string); ok {
				attrs[k] = kv[i+1]
			}
		}
	}

	if c := middleware.LogCollectorFromContext(ctx); c != nil {
		c.Append(store.LogEntry{
			Time:    time.Now(),
			Level:   "EVENT",
			Message: name,
			Attrs:   attrs,
		})
	}
	if tracer := middleware.GetTracer(ctx); tracer != nil {
		tracer.RecordCustom(name, attrs)
	}
}
