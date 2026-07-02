package govisual

import (
	"context"
	"log/slog"

	"github.com/doganarif/govisual/v2/internal/middleware"
	"github.com/doganarif/govisual/v2/store"
)

// SlogHandler wraps a slog.Handler so records logged with a request's
// context also attach to that request in the dashboard:
//
//	logger := slog.New(govisual.SlogHandler(slog.NewJSONHandler(os.Stdout, nil)))
//	logger.InfoContext(r.Context(), "cache miss", "key", key)
//
// Records logged without a captured request's context pass through to the
// base handler untouched. Capture is bounded per request; a floods-of-logs
// handler can't grow a request record without limit.
func SlogHandler(base slog.Handler) slog.Handler {
	return &slogCapture{base: base}
}

type slogCapture struct {
	base  slog.Handler
	attrs []slog.Attr
	group string
}

func (h *slogCapture) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *slogCapture) Handle(ctx context.Context, rec slog.Record) error {
	if c := middleware.LogCollectorFromContext(ctx); c != nil {
		attrs := make(map[string]any, rec.NumAttrs()+len(h.attrs))
		for _, a := range h.attrs {
			attrs[a.Key] = a.Value.Any()
		}
		rec.Attrs(func(a slog.Attr) bool {
			attrs[h.group+a.Key] = a.Value.Any()
			return true
		})
		c.Append(store.LogEntry{
			Time:    rec.Time,
			Level:   rec.Level.String(),
			Message: rec.Message,
			Attrs:   attrs,
		})
	}
	return h.base.Handle(ctx, rec)
}

func (h *slogCapture) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Groups flatten to dotted keys in the captured map.
	prefixed := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		prefixed[i] = slog.Attr{Key: h.group + a.Key, Value: a.Value}
	}
	return &slogCapture{
		base:  h.base.WithAttrs(attrs),
		attrs: append(append([]slog.Attr(nil), h.attrs...), prefixed...),
		group: h.group,
	}
}

func (h *slogCapture) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &slogCapture{
		base:  h.base.WithGroup(name),
		attrs: h.attrs,
		group: h.group + name + ".",
	}
}
