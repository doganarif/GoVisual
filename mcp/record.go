package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// recorded wraps a tool handler so successful and failing calls both land
// on the shared activity log. Bodies aren't included — just an
// abbreviated view of the arguments, so a huge replay body doesn't bloat
// the dashboard.
func recorded[Args any, Result any](
	cfg *config,
	tool string,
	mutating bool,
	fn func(ctx context.Context, req *sdk.CallToolRequest, args Args) (*sdk.CallToolResult, Result, error),
) func(ctx context.Context, req *sdk.CallToolRequest, args Args) (*sdk.CallToolResult, Result, error) {
	if cfg == nil || cfg.activity == nil {
		return fn
	}
	return func(ctx context.Context, req *sdk.CallToolRequest, args Args) (*sdk.CallToolResult, Result, error) {
		start := time.Now()
		res, out, err := fn(ctx, req, args)
		entry := store.ActivityEntry{
			Tool:     tool,
			Args:     summarizeArgs(args),
			Duration: time.Since(start),
			Mutating: mutating,
		}
		if err != nil {
			entry.Error = err.Error()
		} else if res != nil && res.IsError {
			entry.Error = "tool returned an error result"
		}
		cfg.activity.Record(entry)
		return res, out, err
	}
}

// summarizeArgs turns typed tool arguments into a compact string map for
// display. Non-string values are JSON-encoded, long values truncated —
// this feeds a UI list, not an audit log.
func summarizeArgs(v any) map[string]string {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, val := range raw {
		s, ok := val.(string)
		if !ok {
			b, _ := json.Marshal(val)
			s = string(b)
		}
		if len(s) > 120 {
			s = s[:120] + fmt.Sprintf(" …[+%d bytes]", len(s)-120)
		}
		out[k] = s
	}
	return out
}
