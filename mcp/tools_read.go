package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Body and result limits keep a single tool call from flooding an agent's
// context window. Per-call overrides may raise the defaults only to these
// hard ceilings.
const (
	defaultBodyBytes = 2048
	maxBodyBytes     = 64 << 10
	defaultListLimit = 20
	maxListLimit     = 100
	maxDetailEntries = 50
	maxHeaderEntries = 100
	maxAttrEntries   = 20
	maxHeaderBytes   = 4 << 10
	maxTextBytes     = 64 << 10
	maxReportBytes   = 256 << 10
)

type requestSummary struct {
	ID       string `json:"id"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Status   int    `json:"status"`
	Duration int64  `json:"duration_ms"`
	Time     string `json:"time"`
	Error    string `json:"error,omitempty"`
}

func summarize(l *store.RequestLog) requestSummary {
	return requestSummary{
		ID:       truncateField(l.ID, maxHeaderBytes),
		Method:   truncateField(l.Method, maxHeaderBytes),
		Path:     truncateField(l.Path, maxHeaderBytes),
		Status:   l.StatusCode,
		Duration: l.Duration,
		Time:     l.Timestamp.Format("15:04:05.000"),
		Error:    truncateField(l.Error, maxHeaderBytes),
	}
}

type requestDetail struct {
	requestSummary
	Host              string             `json:"host,omitempty"`
	RawPath           string             `json:"raw_path,omitempty"`
	Query             string             `json:"query,omitempty"`
	RequestHeaders    map[string]string  `json:"request_headers,omitempty"`
	ResponseHeaders   map[string]string  `json:"response_headers,omitempty"`
	RequestBody       string             `json:"request_body,omitempty"`
	RequestBodyBytes  int                `json:"request_body_bytes,omitempty"`
	ResponseBody      string             `json:"response_body,omitempty"`
	ResponseBodyBytes int                `json:"response_body_bytes,omitempty"`
	Logs              []store.LogEntry   `json:"logs,omitempty"`
	SQLQueries        []store.SQLQuery   `json:"sql_queries,omitempty"`
	HTTPCalls         []store.HTTPCall   `json:"http_calls,omitempty"`
	Bottlenecks       []store.Bottleneck `json:"bottlenecks,omitempty"`
	PanicStack        string             `json:"panic_stack,omitempty"`
}

func detail(l *store.RequestLog, maxBody int) requestDetail {
	maxBody = bodyLimit(maxBody)
	d := requestDetail{
		requestSummary:    summarize(l),
		Host:              truncateField(l.Host, maxHeaderBytes),
		RawPath:           truncateField(l.RawPath, maxTextBytes),
		Query:             truncateField(l.Query, maxTextBytes),
		RequestHeaders:    flattenHeaders(l.RequestHeaders),
		ResponseHeaders:   flattenHeaders(l.ResponseHeaders),
		RequestBody:       truncate(l.RequestBody, maxBody),
		RequestBodyBytes:  len(l.RequestBody),
		ResponseBody:      truncate(l.ResponseBody, maxBody),
		ResponseBodyBytes: len(l.ResponseBody),
		Logs:              boundedLogs(l.Logs),
		PanicStack:        truncateField(l.PanicStack, maxTextBytes),
	}
	if m := l.PerformanceMetrics; m != nil {
		d.SQLQueries = boundedSQLQueries(m.SQLQueries)
		d.HTTPCalls = boundedHTTPCalls(m.HTTPCalls)
		d.Bottlenecks = boundedBottlenecks(m.Bottlenecks)
	}
	return d
}

func bodyLimit(n int) int {
	if n <= 0 {
		return defaultBodyBytes
	}
	if n > maxBodyBytes {
		return maxBodyBytes
	}
	return n
}

func resultLimit(n int) int {
	if n <= 0 {
		return defaultListLimit
	}
	if n > maxListLimit {
		return maxListLimit
	}
	return n
}

func flattenHeaders(h map[string][]string) map[string]string {
	if len(h) == 0 {
		return nil
	}
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > maxHeaderEntries {
		keys = keys[:maxHeaderEntries]
	}
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[truncateField(key, maxHeaderBytes)] = truncateField(strings.Join(h[key], ", "), maxHeaderBytes)
	}
	return out
}

func boundedLogs(entries []store.LogEntry) []store.LogEntry {
	if len(entries) > maxDetailEntries {
		entries = entries[:maxDetailEntries]
	}
	out := make([]store.LogEntry, 0, len(entries))
	for _, entry := range entries {
		entry.Level = truncateField(entry.Level, maxHeaderBytes)
		entry.Message = truncateField(entry.Message, maxHeaderBytes)
		if len(entry.Attrs) > 0 {
			keys := make([]string, 0, len(entry.Attrs))
			for key := range entry.Attrs {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			if len(keys) > maxAttrEntries {
				keys = keys[:maxAttrEntries]
			}
			attrs := make(map[string]any, len(keys))
			for _, key := range keys {
				attrs[truncateField(key, maxHeaderBytes)] = truncateField(fmt.Sprint(entry.Attrs[key]), maxHeaderBytes)
			}
			entry.Attrs = attrs
		}
		out = append(out, entry)
	}
	return out
}

func boundedSQLQueries(entries []store.SQLQuery) []store.SQLQuery {
	if len(entries) > maxDetailEntries {
		entries = entries[:maxDetailEntries]
	}
	out := append([]store.SQLQuery(nil), entries...)
	for i := range out {
		out[i].Query = truncateField(out[i].Query, maxHeaderBytes)
		out[i].Error = truncateField(out[i].Error, maxHeaderBytes)
	}
	return out
}

func boundedHTTPCalls(entries []store.HTTPCall) []store.HTTPCall {
	if len(entries) > maxDetailEntries {
		entries = entries[:maxDetailEntries]
	}
	out := append([]store.HTTPCall(nil), entries...)
	for i := range out {
		out[i].Method = truncateField(out[i].Method, maxHeaderBytes)
		out[i].URL = truncateField(out[i].URL, maxHeaderBytes)
	}
	return out
}

func boundedBottlenecks(entries []store.Bottleneck) []store.Bottleneck {
	if len(entries) > maxDetailEntries {
		entries = entries[:maxDetailEntries]
	}
	out := append([]store.Bottleneck(nil), entries...)
	for i := range out {
		out[i].Type = truncateField(out[i].Type, maxHeaderBytes)
		out[i].Description = truncateField(out[i].Description, maxHeaderBytes)
		out[i].Suggestion = truncateField(out[i].Suggestion, maxHeaderBytes)
	}
	return out
}

func truncateField(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + fmt.Sprintf("... [%d more bytes]", len(value)-max)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + fmt.Sprintf("... [%d more bytes; raise max_body_bytes to see more]", len(s)-n)
}

func isFailure(l *store.RequestLog) bool {
	return l.StatusCode >= 400 || l.Error != ""
}

type listArgs struct {
	Limit        int    `json:"limit,omitempty"`
	Method       string `json:"method,omitempty"`
	PathContains string `json:"path_contains,omitempty"`
	StatusMin    int    `json:"status_min,omitempty"`
	StatusMax    int    `json:"status_max,omitempty"`
	ErrorsOnly   bool   `json:"errors_only,omitempty"`
}

type listResult struct {
	Requests []requestSummary `json:"requests"`
	Total    int              `json:"total_in_store"`
	Note     string           `json:"note,omitempty"`
}

type idArgs struct {
	ID           string `json:"id"`
	MaxBodyBytes int    `json:"max_body_bytes,omitempty"`
}

type searchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type emptyArgs struct{}

type statsResult struct {
	Routes      []routeStats `json:"routes"`
	TotalRoutes int          `json:"total_routes"`
	Note        string       `json:"note,omitempty"`
}

type reportResult struct {
	Report string `json:"report"`
}

type routeStats struct {
	Route      string `json:"route"`
	Count      int    `json:"count"`
	Errors     int    `json:"errors"`
	P50Millis  int64  `json:"p50_ms"`
	P95Millis  int64  `json:"p95_ms"`
	LastSeenAt string `json:"last_seen"`
}

func registerReadTools(srv *sdk.Server, st store.Store, cfg *config) {
	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_last_error",
		Description: "Return the most recent failed request (status >= 400 or a recorded error/panic) " +
			"with headers, body excerpts, logs, SQL queries and panic stack. Start here when debugging.",
	}, recorded(cfg, "get_last_error", false, func(ctx context.Context, req *sdk.CallToolRequest, args emptyArgs) (*sdk.CallToolResult, requestDetail, error) {
		for _, l := range st.GetLatest(200) {
			if isFailure(l) {
				return nil, detail(l, defaultBodyBytes), nil
			}
		}
		return nil, requestDetail{}, fmt.Errorf("no failed requests captured; %d requests in store", len(st.GetAll()))
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_requests",
		Description: "List captured requests, newest first. Filters: method, path_contains, status_min/status_max, " +
			"errors_only. limit defaults to 20. Returns compact summaries; use get_request for detail.",
	}, recorded(cfg, "list_requests", false, func(ctx context.Context, req *sdk.CallToolRequest, args listArgs) (*sdk.CallToolResult, listResult, error) {
		limit := resultLimit(args.Limit)
		all := st.GetAll()
		out := make([]requestSummary, 0, limit)
		for _, l := range all {
			if args.Method != "" && !strings.EqualFold(args.Method, l.Method) {
				continue
			}
			if args.PathContains != "" && !strings.Contains(l.Path, args.PathContains) {
				continue
			}
			if args.StatusMin > 0 && l.StatusCode < args.StatusMin {
				continue
			}
			if args.StatusMax > 0 && l.StatusCode > args.StatusMax {
				continue
			}
			if args.ErrorsOnly && !isFailure(l) {
				continue
			}
			out = append(out, summarize(l))
			if len(out) == limit {
				break
			}
		}
		res := listResult{Requests: out, Total: len(all)}
		if len(out) == limit && len(all) > limit {
			res.Note = "more matches may exist; refine filters or raise limit up to 100"
		}
		return nil, res, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_request",
		Description: "Full detail for one captured request by id: headers, body excerpts (capped by " +
			"max_body_bytes, default 2048), logs, SQL queries, outbound HTTP calls, bottlenecks, panic stack.",
	}, recorded(cfg, "get_request", false, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, requestDetail, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, requestDetail{}, fmt.Errorf("no request with id %q; it may have rolled out of the store", args.ID)
		}
		return nil, detail(l, args.MaxBodyBytes), nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "search_requests",
		Description: "Substring search across path, query string, request/response bodies, and errors. " +
			"Returns compact summaries, newest first.",
	}, recorded(cfg, "search_requests", false, func(ctx context.Context, req *sdk.CallToolRequest, args searchArgs) (*sdk.CallToolResult, listResult, error) {
		if args.Query == "" {
			return nil, listResult{}, fmt.Errorf("query is required")
		}
		limit := resultLimit(args.Limit)
		all := st.GetAll()
		out := make([]requestSummary, 0, limit)
		for _, l := range all {
			if !matches(l, args.Query) {
				continue
			}
			out = append(out, summarize(l))
			if len(out) == limit {
				break
			}
		}
		res := listResult{Requests: out, Total: len(all)}
		if len(out) == limit && len(all) > limit {
			res.Note = "more matches may exist; refine the query or raise limit up to 100"
		}
		return nil, res, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_stats",
		Description: "Aggregate view of captured traffic: per-route request count, error count, p50/p95 " +
			"latency. Use this to find slow or failing routes before drilling in.",
	}, recorded(cfg, "get_stats", false, func(ctx context.Context, req *sdk.CallToolRequest, args emptyArgs) (*sdk.CallToolResult, statsResult, error) {
		byRoute := map[string][]*store.RequestLog{}
		for _, l := range st.GetAll() {
			key := l.Method + " " + l.Path
			byRoute[key] = append(byRoute[key], l)
		}
		out := make([]routeStats, 0, len(byRoute))
		for route, logs := range byRoute {
			durations := make([]int64, 0, len(logs))
			errors := 0
			last := logs[0].Timestamp
			for _, l := range logs {
				durations = append(durations, l.Duration)
				if isFailure(l) {
					errors++
				}
				if l.Timestamp.After(last) {
					last = l.Timestamp
				}
			}
			sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
			out = append(out, routeStats{
				Route:      truncateField(route, maxHeaderBytes),
				Count:      len(logs),
				Errors:     errors,
				P50Millis:  percentile(durations, 50),
				P95Millis:  percentile(durations, 95),
				LastSeenAt: last.Format("15:04:05"),
			})
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].Count == out[j].Count {
				return out[i].Route < out[j].Route
			}
			return out[i].Count > out[j].Count
		})
		result := statsResult{TotalRoutes: len(out)}
		if len(out) > maxListLimit {
			result.Routes = out[:maxListLimit]
			result.Note = "showing the 100 busiest routes"
		} else {
			result.Routes = out
		}
		return nil, result, nil
	}))

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_debug_context",
		Description: "Everything known about one request as a single readable report: request line, status, " +
			"timing, headers, body excerpts, application logs, SQL queries, outbound calls, and panic stack. " +
			"The one-call version of get_request for when you want the whole crime scene.",
	}, recorded(cfg, "get_debug_context", false, func(ctx context.Context, req *sdk.CallToolRequest, args idArgs) (*sdk.CallToolResult, reportResult, error) {
		l, ok := st.Get(args.ID)
		if !ok {
			return nil, reportResult{}, fmt.Errorf("no request with id %q", args.ID)
		}
		return nil, reportResult{Report: debugReport(l, args.MaxBodyBytes)}, nil
	}))
}

func matches(l *store.RequestLog, q string) bool {
	q = strings.ToLower(q)
	for _, s := range []string{l.Path, l.Query, l.RequestBody, l.ResponseBody, l.Error} {
		if strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := len(sorted) * p / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func debugReport(l *store.RequestLog, maxBody int) string {
	maxBody = bodyLimit(maxBody)
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s%s -> %d (%dms) at %s\n", truncateField(l.Method, maxHeaderBytes), truncateField(l.Path, maxTextBytes), querySuffix(truncateField(l.Query, maxTextBytes)), l.StatusCode, l.Duration, l.Timestamp.Format("15:04:05.000"))
	if l.Error != "" {
		fmt.Fprintf(&b, "ERROR: %s\n", truncateField(l.Error, maxTextBytes))
	}
	if len(l.RequestHeaders) > 0 {
		b.WriteString("\nRequest headers:\n")
		writeHeaders(&b, l.RequestHeaders)
	}
	if l.RequestBody != "" {
		fmt.Fprintf(&b, "\nRequest body (%d bytes):\n%s\n", len(l.RequestBody), truncate(l.RequestBody, maxBody))
	}
	if l.ResponseBody != "" {
		fmt.Fprintf(&b, "\nResponse body (%d bytes):\n%s\n", len(l.ResponseBody), truncate(l.ResponseBody, maxBody))
	}
	if len(l.Logs) > 0 {
		b.WriteString("\nApplication logs:\n")
		for _, e := range boundedLogs(l.Logs) {
			fmt.Fprintf(&b, "  %s [%s] %s", e.Time.Format("15:04:05.000"), e.Level, e.Message)
			for k, v := range e.Attrs {
				fmt.Fprintf(&b, " %s=%v", k, v)
			}
			b.WriteString("\n")
		}
	}
	if m := l.PerformanceMetrics; m != nil {
		if len(m.SQLQueries) > 0 {
			fmt.Fprintf(&b, "\nSQL queries (%d):\n", len(m.SQLQueries))
			for _, q := range boundedSQLQueries(m.SQLQueries) {
				fmt.Fprintf(&b, "  %s (%s, %d rows)", truncateField(q.Query, 200), q.Duration, q.Rows)
				if q.Error != "" {
					fmt.Fprintf(&b, " ERROR: %s", q.Error)
				}
				b.WriteString("\n")
			}
		}
		if len(m.HTTPCalls) > 0 {
			fmt.Fprintf(&b, "\nOutbound HTTP calls (%d):\n", len(m.HTTPCalls))
			for _, c := range boundedHTTPCalls(m.HTTPCalls) {
				fmt.Fprintf(&b, "  %s %s -> %d (%s)\n", c.Method, c.URL, c.Status, c.Duration)
			}
		}
		for _, bn := range boundedBottlenecks(m.Bottlenecks) {
			fmt.Fprintf(&b, "\nBottleneck [%s]: %s — %s\n", bn.Type, bn.Description, bn.Suggestion)
		}
	}
	if l.PanicStack != "" {
		fmt.Fprintf(&b, "\nPanic stack:\n%s\n", truncateField(l.PanicStack, maxTextBytes))
	}
	return truncateField(b.String(), maxReportBytes)
}

func querySuffix(q string) string {
	if q == "" {
		return ""
	}
	return "?" + q
}

func writeHeaders(b *strings.Builder, h map[string][]string) {
	headers := flattenHeaders(h)
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := headers[key]
		fmt.Fprintf(b, "  %s: %s\n", key, value)
	}
}
