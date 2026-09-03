package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// summarizeArgs turns typed tool arguments into a compact, redacted string map
// for display. Request bodies are omitted and credential-shaped values are
// redacted before any nested value is encoded.
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
		val = sanitizeActivityArg(k, val)
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

func sanitizeActivityArg(key string, value any) any {
	normalized := strings.NewReplacer("-", "", "_", "").Replace(strings.ToLower(key))
	if normalized == "body" || strings.HasSuffix(normalized, "body") {
		return "[omitted]"
	}
	if strings.HasSuffix(normalized, "authorization") || normalized == "cookie" || normalized == "setcookie" ||
		strings.Contains(normalized, "apikey") || strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") {
		return "[redacted]"
	}
	switch typed := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(typed))
		for nestedKey, nestedValue := range typed {
			clean[nestedKey] = sanitizeActivityArg(nestedKey, nestedValue)
		}
		return clean
	case []any:
		clean := make([]any, len(typed))
		for i, nestedValue := range typed {
			clean[i] = sanitizeActivityArg(key, nestedValue)
		}
		return clean
	default:
		return value
	}
}
