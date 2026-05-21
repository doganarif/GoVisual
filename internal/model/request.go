package model

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/profiling"
)

type RequestLog struct {
	ID                 string                   `json:"ID" bson:"_id"`
	Timestamp          time.Time                `json:"Timestamp" bson:"timestamp"`
	Method             string                   `json:"Method" bson:"method"`
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
	PerformanceMetrics *profiling.Metrics       `json:"PerformanceMetrics,omitempty" bson:"performance_metrics,omitempty"`
}

func NewRequestLog(req *http.Request) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		Method:         req.Method,
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
