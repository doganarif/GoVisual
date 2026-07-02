package store

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

type RequestLog struct {
	ID                 string                   `json:"ID" bson:"_id"`
	Timestamp          time.Time                `json:"Timestamp" bson:"timestamp"`
	Method             string                   `json:"Method" bson:"method"`
	Host               string                   `json:"Host,omitempty" bson:"host,omitempty"`
	Path               string                   `json:"Path" bson:"path"`
	Query              string                   `json:"Query" bson:"query"`
	RequestHeaders     http.Header              `json:"RequestHeaders" bson:"request_headers"`
	ResponseHeaders    http.Header              `json:"ResponseHeaders" bson:"response_headers"`
	StatusCode         int                      `json:"StatusCode" bson:"status_code"`
	Duration           int64                    `json:"Duration" bson:"duration"`
	RequestBody        string                   `json:"RequestBody,omitempty" bson:"request_body,omitempty"`
	ResponseBody       string                   `json:"ResponseBody,omitempty" bson:"response_body,omitempty"`
	Error              string                   `json:"Error,omitempty" bson:"error,omitempty"`
	MiddlewareTrace    []map[string]interface{} `json:"MiddlewareTrace,omitempty" bson:"middleware_trace,omitempty"`
	RouteTrace         map[string]interface{}   `json:"RouteTrace,omitempty" bson:"route_trace,omitempty"`
	PerformanceMetrics *PerformanceMetrics      `json:"PerformanceMetrics,omitempty" bson:"performance_metrics,omitempty"`
	Logs               []LogEntry               `json:"Logs,omitempty" bson:"logs,omitempty"`
	PanicStack         string                   `json:"PanicStack,omitempty" bson:"panic_stack,omitempty"`
}

// LogEntry is a single application log line emitted while the request was
// being handled, captured via govisual.SlogHandler.
type LogEntry struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

// PerformanceMetrics is the per-request profiling data attached to a
// RequestLog. It is plain data; the profiler that produces it lives in
// internal/profiling.
type PerformanceMetrics struct {
	RequestID        string                   `json:"request_id"`
	StartTime        time.Time                `json:"start_time"`
	EndTime          time.Time                `json:"end_time"`
	Duration         time.Duration            `json:"duration"`
	CPUTime          time.Duration            `json:"cpu_time"`
	MemoryAlloc      uint64                   `json:"memory_alloc"`
	MemoryTotalAlloc uint64                   `json:"memory_total_alloc"`
	NumGoroutines    int                      `json:"num_goroutines"`
	NumGC            uint32                   `json:"num_gc"`
	GCPauseTotal     time.Duration            `json:"gc_pause_total"`
	FunctionTimings  map[string]time.Duration `json:"function_timings,omitempty"`
	SQLQueries       []SQLQuery               `json:"sql_queries,omitempty"`
	HTTPCalls        []HTTPCall               `json:"http_calls,omitempty"`
	Bottlenecks      []Bottleneck             `json:"bottlenecks,omitempty"`
	CPUProfile       []byte                   `json:"-"`
	HeapProfile      []byte                   `json:"-"`
}

// SQLQuery is a single captured database query.
type SQLQuery struct {
	Query    string        `json:"query"`
	Duration time.Duration `json:"duration"`
	Rows     int           `json:"rows"`
	Error    string        `json:"error,omitempty"`
}

// HTTPCall is a single captured outbound HTTP call.
type HTTPCall struct {
	Method   string        `json:"method"`
	URL      string        `json:"url"`
	Duration time.Duration `json:"duration"`
	Status   int           `json:"status"`
	Size     int64         `json:"size"`
}

// Bottleneck is a performance problem detected on a request.
type Bottleneck struct {
	Type        string        `json:"type"` // "cpu", "memory", "io", "database", "http"
	Description string        `json:"description"`
	Impact      float64       `json:"impact"` // 0-1 scale of impact
	Duration    time.Duration `json:"duration"`
	Suggestion  string        `json:"suggestion"`
}

func NewRequestLog(req *http.Request) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		Method:         req.Method,
		Host:           req.Host,
		Path:           req.URL.Path,
		Query:          req.URL.RawQuery,
		RequestHeaders: scrubHeaders(req.Header),
	}
}

// sensitiveHeaders are dropped from captured request/response logs. Storing
// raw credentials makes the dashboard a high-value target and creates a
// data-at-rest liability on every configured backend; opt-out is not offered
// because there is no defensible reason to log a bearer token verbatim.
var sensitiveHeaders = map[string]struct{}{
	"Authorization":       {},
	"Proxy-Authorization": {},
	"Cookie":              {},
	"Set-Cookie":          {},
	"X-Api-Key":           {},
	"X-Auth-Token":        {},
	"X-Csrf-Token":        {},
}

// scrubHeaders returns a copy of h with credential-bearing header values
// replaced by a fixed marker. The header *name* is kept so consumers can see
// that auth was present; only the value is hidden.
func scrubHeaders(h http.Header) http.Header {
	if len(h) == 0 {
		return h
	}
	out := make(http.Header, len(h))
	for k, vs := range h {
		if _, redact := sensitiveHeaders[http.CanonicalHeaderKey(k)]; redact {
			out[k] = []string{"[redacted by govisual]"}
			continue
		}
		out[k] = append([]string(nil), vs...)
	}
	return out
}

// generateID returns a collision-resistant 128-bit random identifier
// encoded as 32 hex characters. Falls back to nanosecond timestamp
// only if the OS RNG is unavailable, which should never happen in practice.
func generateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b[:])
}
