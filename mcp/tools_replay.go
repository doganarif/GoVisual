package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// replayClient deliberately has no redirect-following: a replay should show
// what the app returned, not where a 302 leads.
var replayClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type replayArgs struct {
	ID      string            `json:"id"`
	Method  string            `json:"method,omitempty"`
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type replayResult struct {
	Status     int    `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	Body       string `json:"body"`
	BodyBytes  int    `json:"body_bytes"`
}

type diffResult struct {
	OriginalStatus int    `json:"original_status"`
	ReplayStatus   int    `json:"replay_status"`
	StatusChanged  bool   `json:"status_changed"`
	BodyChanged    bool   `json:"body_changed"`
	OriginalMS     int64  `json:"original_ms"`
	ReplayMS       int64  `json:"replay_ms"`
	ReplayBody     string `json:"replay_body,omitempty"`
	Summary        string `json:"summary"`
}

func registerReplayTools(srv *sdk.Server, st store.Store, cfg *config) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "replay_request",
		Description: "Re-send a captured request against the application, optionally overriding method, path, " +
			"headers, or body. The destination host is fixed to the application — only the request shape is " +
			"yours to change. Returns status, duration and a body excerpt.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args replayArgs) (*sdk.CallToolResult, replayResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, replayResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		res, err := replay(ctx, cfg, l, args)
		if err != nil {
			return nil, replayResult{}, err
		}
		return nil, *res, nil
	})

	sdk.AddTool(srv, &sdk.Tool{
		Name: "diff_replay",
		Description: "Replay a captured request unchanged against the current code and diff the outcome " +
			"against the original capture: status, body, timing. Use after changing code to verify a fix " +
			"or check for regressions.",
	}, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, diffResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, diffResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		res, err := replay(ctx, cfg, l, replayArgs{ID: args.ID})
		if err != nil {
			return nil, diffResult{}, err
		}

		d := diffResult{
			OriginalStatus: l.StatusCode,
			ReplayStatus:   res.Status,
			StatusChanged:  l.StatusCode != res.Status,
			BodyChanged:    bodyChanged(l.ResponseBody, res.Body),
			OriginalMS:     l.Duration,
			ReplayMS:       res.DurationMS,
		}
		switch {
		case d.StatusChanged:
			d.Summary = fmt.Sprintf("status changed: %d -> %d", d.OriginalStatus, d.ReplayStatus)
			d.ReplayBody = truncate(res.Body, defaultBodyBytes)
		case d.BodyChanged:
			d.Summary = "status unchanged, body differs"
			d.ReplayBody = truncate(res.Body, defaultBodyBytes)
		default:
			d.Summary = "no change: same status, same body"
		}
		if l.ResponseBody == "" && res.Body != "" {
			d.Summary += " (original capture has no response body; enable WithResponseBodyLogging for real body diffs)"
		}
		return nil, d, nil
	})
}

func replay(ctx context.Context, cfg *config, l *store.RequestLog, args replayArgs) (*replayResult, error) {
	base := cfg.baseURL
	if base == "" {
		if l.Host == "" {
			return nil, fmt.Errorf("no replay base URL: configure WithBaseURL or capture requests with a Host")
		}
		base = "http://" + l.Host
	}

	method := l.Method
	if args.Method != "" {
		method = args.Method
	}
	path := l.Path
	if args.Path != "" {
		path = args.Path
	}
	url := strings.TrimSuffix(base, "/") + path
	if l.Query != "" && args.Path == "" {
		url += "?" + l.Query
	}

	body := l.RequestBody
	if args.Body != "" {
		body = args.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building replay request: %w", err)
	}
	for k, vs := range l.RequestHeaders {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("X-Govisual-Replay", "1")

	start := time.Now()
	resp, err := replayClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("replay failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading replay response: %w", err)
	}

	return &replayResult{
		Status:     resp.StatusCode,
		DurationMS: time.Since(start).Milliseconds(),
		Body:       truncate(string(respBody), defaultBodyBytes),
		BodyBytes:  len(respBody),
	}, nil
}

// bodyChanged compares bodies structurally when both parse as JSON, so key
// order and whitespace don't count as regressions.
func bodyChanged(original, replayed string) bool {
	if original == replayed {
		return false
	}
	var a, b any
	if json.Unmarshal([]byte(original), &a) == nil && json.Unmarshal([]byte(replayed), &b) == nil {
		return !reflect.DeepEqual(a, b)
	}
	return true
}
