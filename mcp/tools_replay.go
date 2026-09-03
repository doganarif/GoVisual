package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxReplayBodyBytes = 1 << 20
const redactedHeaderValue = "[redacted by govisual]"
const captureTruncationMarker = "...[truncated by govisual]"

// replayClient deliberately has no redirect-following: a replay should show
// what the app returned, not where a 302 leads. DisableCompression preserves
// the captured Accept-Encoding request semantics and the exact response bytes.
var replayClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: &http.Transport{Proxy: nil, DisableCompression: true},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type replayArgs struct {
	ID      string            `json:"id"`
	Method  string            `json:"method,omitempty"`
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	// RemoveHeaders removes captured headers before applying overrides.
	RemoveHeaders []string `json:"remove_headers,omitempty"`
	Body          *string  `json:"body,omitempty"`
}

type replayResult struct {
	Status         int      `json:"status"`
	DurationMS     int64    `json:"duration_ms"`
	Body           string   `json:"body"`
	BodyBytes      int      `json:"body_bytes"`
	BodyTruncated  bool     `json:"body_truncated,omitempty"`
	OmittedHeaders []string `json:"omitted_redacted_headers,omitempty"`
	fullBody       string
	truncated      bool
}

type diffReplayArgs struct {
	ID            string            `json:"id"`
	Headers       map[string]string `json:"headers,omitempty"`
	RemoveHeaders []string          `json:"remove_headers,omitempty"`
	Body          *string           `json:"body,omitempty"`
}

type diffResult struct {
	OriginalStatus int      `json:"original_status"`
	ReplayStatus   int      `json:"replay_status"`
	StatusChanged  bool     `json:"status_changed"`
	BodyChanged    bool     `json:"body_changed"`
	OriginalMS     int64    `json:"original_ms"`
	ReplayMS       int64    `json:"replay_ms"`
	ReplayBody     string   `json:"replay_body,omitempty"`
	BodyCompared   bool     `json:"body_compared"`
	BodyTruncated  bool     `json:"body_comparison_truncated,omitempty"`
	OmittedHeaders []string `json:"omitted_redacted_headers,omitempty"`
	Summary        string   `json:"summary"`
}

func registerReplayTools(srv *sdk.Server, st store.Store, cfg *config) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "replay_request",
		Description: "Re-send a captured request against the application, optionally overriding method, path, " +
			"headers, or body. The destination host is fixed to the application — only the request shape is " +
			"yours to change. Returns status, duration and a body excerpt.",
	}, recorded(cfg, "replay_request", true, func(ctx context.Context, req *sdk.CallToolRequest, args replayArgs) (*sdk.CallToolResult, replayResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, replayResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		res, err := replay(ctx, cfg, l, args)
		if err != nil {
			return nil, replayResult{}, err
		}
		return nil, *res, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "diff_replay",
		Description: "Replay a captured request unchanged against the current code and diff the outcome " +
			"against the original capture: status, body, timing. Use after changing code to verify a fix " +
			"or check for regressions.",
	}, recorded(cfg, "diff_replay", true, func(ctx context.Context, req *sdk.CallToolRequest, args diffReplayArgs) (*sdk.CallToolResult, diffResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, diffResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		res, err := replay(ctx, cfg, l, replayArgs{ID: args.ID, Headers: args.Headers, RemoveHeaders: args.RemoveHeaders, Body: args.Body})
		if err != nil {
			return nil, diffResult{}, err
		}

		originalBody, originalTruncated, bodyAvailable := comparableCapturedBody(l.ResponseBody)
		d := diffResult{
			OriginalStatus: l.StatusCode,
			ReplayStatus:   res.Status,
			StatusChanged:  l.StatusCode != res.Status,
			BodyCompared:   bodyAvailable,
			BodyTruncated:  originalTruncated || res.truncated,
			OmittedHeaders: res.OmittedHeaders,
			OriginalMS:     l.Duration,
			ReplayMS:       res.DurationMS,
		}
		if bodyAvailable {
			d.BodyChanged = bodyChanged(originalBody, comparableReplayBody(res.fullBody, len(originalBody), originalTruncated))
		}
		switch {
		case d.StatusChanged:
			d.Summary = fmt.Sprintf("status changed: %d -> %d", d.OriginalStatus, d.ReplayStatus)
			d.ReplayBody = truncate(res.fullBody, defaultBodyBytes)
		case d.BodyChanged:
			d.Summary = "status unchanged, body differs"
			d.ReplayBody = truncate(res.fullBody, defaultBodyBytes)
		case !d.BodyCompared:
			d.Summary = "status unchanged; original response body was not captured, so body comparison is unavailable"
		case d.BodyTruncated:
			d.Summary = "status unchanged; captured body prefix matches, but comparison was truncated"
		default:
			d.Summary = "no change: same status, same body"
		}
		if len(d.OmittedHeaders) > 0 {
			d.Summary += "; redacted request headers were omitted (provide header overrides for an authenticated comparison)"
		}
		return nil, d, nil
	}))
}

func replay(ctx context.Context, cfg *config, l *store.RequestLog, args replayArgs) (*replayResult, error) {
	method := l.Method
	if args.Method != "" {
		method = args.Method
	}
	baseURL, err := replayBaseURL(cfg, l)
	if err != nil {
		return nil, err
	}
	target, err := replayTarget(baseURL, l.Path, l.RawPath, l.Query, args.Path)
	if err != nil {
		return nil, err
	}

	body := l.RequestBody
	if args.Body == nil && strings.HasSuffix(body, captureTruncationMarker) {
		return nil, fmt.Errorf("captured request body is truncated; provide the complete body override before replaying")
	}
	if args.Body != nil {
		body = *args.Body
	}
	if len(body) > maxReplayBodyBytes {
		return nil, fmt.Errorf("replay request body exceeds %d bytes", maxReplayBodyBytes)
	}
	if len(method) > maxHeaderBytes || len(target) > maxTextBytes {
		return nil, fmt.Errorf("replay method or target is too large")
	}

	req, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building replay request: %w", err)
	}
	copyReplayHeaders(req.Header, l.RequestHeaders, args.Headers, args.RemoveHeaders)
	req.Header.Set("X-Govisual-Replay", "1")
	if err := validateReplayHeaders(req.Header); err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := replayClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("replay failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading replay response: %w", err)
	}

	truncated := len(respBody) > maxReplayBodyBytes
	if truncated {
		respBody = respBody[:maxReplayBodyBytes]
	}
	return &replayResult{
		Status:         resp.StatusCode,
		DurationMS:     time.Since(start).Milliseconds(),
		Body:           truncate(string(respBody), defaultBodyBytes),
		BodyBytes:      len(respBody),
		BodyTruncated:  truncated,
		OmittedHeaders: omittedRedactedHeaders(l.RequestHeaders, req.Header),
		fullBody:       string(respBody),
		truncated:      truncated,
	}, nil
}

func validateReplayHeaders(headers http.Header) error {
	if len(headers) > maxHeaderEntries {
		return fmt.Errorf("replay has too many headers (maximum %d)", maxHeaderEntries)
	}
	remaining := maxGeneratedBytes
	for key, values := range headers {
		if len(key) > maxHeaderBytes || len(key) > remaining {
			return fmt.Errorf("replay headers exceed size limits")
		}
		remaining -= len(key)
		for _, value := range values {
			if len(value) > maxTextBytes || len(value) > remaining {
				return fmt.Errorf("replay headers exceed size limits")
			}
			remaining -= len(value)
		}
	}
	return nil
}

func replayBaseURL(cfg *config, _ *store.RequestLog) (string, error) {
	if cfg != nil && strings.TrimSpace(cfg.baseURL) != "" {
		return cfg.baseURL, nil
	}
	return "", fmt.Errorf("no safe replay destination: configure WithBaseURL")
}

func replayTarget(baseRaw, capturedPath, capturedRawPath, capturedQuery, override string) (string, error) {
	if strings.TrimSpace(baseRaw) == "" {
		return "", fmt.Errorf("no replay base URL: configure WithBaseURL")
	}
	base, err := url.Parse(baseRaw)
	if err != nil || base == nil {
		return "", fmt.Errorf("invalid replay base URL")
	}
	base.Scheme = strings.ToLower(base.Scheme)
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil || base.Opaque != "" || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" {
		return "", fmt.Errorf("invalid replay base URL")
	}
	if _, _, err := canonicalAuthority(base.Host, base.Scheme); err != nil {
		return "", fmt.Errorf("invalid replay base URL")
	}

	requestPath := capturedPath
	requestRawPath := capturedRawPath
	rawQuery := capturedQuery
	if override != "" {
		relative, err := url.Parse(override)
		if err != nil || relative.IsAbs() || relative.Host != "" || relative.Fragment != "" || !strings.HasPrefix(relative.Path, "/") {
			return "", fmt.Errorf("replay path must be an absolute path without a host")
		}
		requestPath = relative.Path
		requestRawPath = relative.RawPath
		rawQuery = relative.RawQuery
	}
	if requestPath == "" || !strings.HasPrefix(requestPath, "/") {
		return "", fmt.Errorf("captured request has an invalid path")
	}

	target := *base
	target.Path = strings.TrimSuffix(base.Path, "/") + requestPath
	if requestRawPath != "" {
		target.RawPath = strings.TrimSuffix(base.EscapedPath(), "/") + requestRawPath
	} else {
		target.RawPath = ""
	}
	target.RawQuery = rawQuery
	return target.String(), nil
}

var fixedReplayHeaders = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Proxy-Connection":    {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
	"Host":                {},
	"Content-Length":      {},
}

func copyReplayHeaders(dst http.Header, captured http.Header, overrides map[string]string, remove []string) {
	blocked := make(map[string]struct{}, len(fixedReplayHeaders))
	for name := range fixedReplayHeaders {
		blocked[name] = struct{}{}
	}
	for key, values := range captured {
		if http.CanonicalHeaderKey(key) == "Connection" {
			blockConnectionTokens(blocked, values...)
		}
	}
	for key, value := range overrides {
		if http.CanonicalHeaderKey(key) == "Connection" {
			blockConnectionTokens(blocked, value)
		}
	}
	for _, key := range remove {
		if canonical := http.CanonicalHeaderKey(key); canonical != "" {
			blocked[canonical] = struct{}{}
		}
	}
	for key, values := range captured {
		canonical := http.CanonicalHeaderKey(key)
		if _, skip := blocked[canonical]; skip || canonical == "" {
			continue
		}
		for _, value := range values {
			if value != redactedHeaderValue {
				dst.Add(canonical, value)
			}
		}
	}
	for key, value := range overrides {
		canonical := http.CanonicalHeaderKey(key)
		if _, skip := blocked[canonical]; skip || canonical == "" || value == redactedHeaderValue {
			continue
		}
		dst.Set(canonical, value)
	}
}

func omittedRedactedHeaders(captured, sent http.Header) []string {
	var omitted []string
	for key, values := range captured {
		canonical := http.CanonicalHeaderKey(key)
		if len(sent.Values(canonical)) > 0 {
			continue
		}
		for _, value := range values {
			if value == redactedHeaderValue {
				omitted = append(omitted, canonical)
				break
			}
		}
	}
	sort.Strings(omitted)
	return omitted
}

func blockConnectionTokens(blocked map[string]struct{}, values ...string) {
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			if canonical := http.CanonicalHeaderKey(strings.TrimSpace(token)); canonical != "" {
				blocked[canonical] = struct{}{}
			}
		}
	}
}

func comparableCapturedBody(body string) (string, bool, bool) {
	if body == "" {
		return "", false, false
	}
	truncated := false
	if strings.HasSuffix(body, captureTruncationMarker) {
		body = strings.TrimSuffix(body, captureTruncationMarker)
		truncated = true
	}
	if len(body) > maxReplayBodyBytes {
		body = body[:maxReplayBodyBytes]
		truncated = true
	}
	return body, truncated, true
}

func comparableReplayBody(body string, originalBytes int, prefixOnly bool) string {
	if prefixOnly && len(body) > originalBytes {
		return body[:originalBytes]
	}
	return body
}

// bodyChanged compares bodies structurally when both parse as JSON, so key
// order and whitespace don't count as regressions.
func bodyChanged(original, replayed string) bool {
	if original == replayed {
		return false
	}
	a, aOK := decodeJSONValue(original)
	b, bOK := decodeJSONValue(replayed)
	if aOK && bOK {
		return !reflect.DeepEqual(a, b)
	}
	return true
}

func decodeJSONValue(raw string) (any, bool) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, false
	}
	return value, true
}
