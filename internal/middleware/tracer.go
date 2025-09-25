package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// TraceEntry represents a single trace entry in the middleware stack
type TraceEntry struct {
	Name      string        `json:"name"`
	Type      string        `json:"type"` // "middleware", "handler", "sql", "http", "custom"
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"` // "running", "completed", "error"
	Error     string        `json:"error,omitempty"`
	Details   interface{}   `json:"details,omitempty"`
	Children  []TraceEntry  `json:"children,omitempty"`
}

// RequestTracer tracks execution through middleware stack
type RequestTracer struct {
	mu          sync.Mutex
	RequestID   string       `json:"request_id"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	Traces      []TraceEntry `json:"traces"`
	currentPath []int        // Stack of indices for nested traces
}

// NewRequestTracer creates a new request tracer
func NewRequestTracer(requestID string) *RequestTracer {
	return &RequestTracer{
		RequestID:   requestID,
		StartTime:   time.Now(),
		Traces:      make([]TraceEntry, 0),
		currentPath: make([]int, 0),
	}
}

// StartTrace starts a new trace entry
func (rt *RequestTracer) StartTrace(name, traceType string, details interface{}) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	entry := TraceEntry{
		Name:      name,
		Type:      traceType,
		StartTime: time.Now(),
		Status:    "running",
		Details:   details,
		Children:  make([]TraceEntry, 0),
	}

	if len(rt.currentPath) == 0 {
		// Root level trace
		rt.Traces = append(rt.Traces, entry)
		rt.currentPath = append(rt.currentPath, len(rt.Traces)-1)
	} else {
		// Nested trace
		parent := rt.getParentTrace()
		if parent != nil {
			parent.Children = append(parent.Children, entry)
			rt.currentPath = append(rt.currentPath, len(parent.Children)-1)
		}
	}
}

// EndTrace ends the current trace entry
func (rt *RequestTracer) EndTrace(err error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if len(rt.currentPath) == 0 {
		return
	}

	trace := rt.getCurrentTrace()
	if trace != nil {
		trace.EndTime = time.Now()
		trace.Duration = trace.EndTime.Sub(trace.StartTime)
		if err != nil {
			trace.Status = "error"
			trace.Error = err.Error()
		} else {
			trace.Status = "completed"
		}
	}

	// Pop from path
	rt.currentPath = rt.currentPath[:len(rt.currentPath)-1]
}

// RecordSQL records a SQL query execution
func (rt *RequestTracer) RecordSQL(query string, duration time.Duration, rows int, err error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	details := map[string]interface{}{
		"query": query,
		"rows":  rows,
	}

	entry := TraceEntry{
		Name:      "SQL Query",
		Type:      "sql",
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Duration:  duration,
		Status:    "completed",
		Details:   details,
	}

	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
	}

	if len(rt.currentPath) > 0 {
		parent := rt.getCurrentTrace()
		if parent != nil {
			parent.Children = append(parent.Children, entry)
		}
	} else {
		rt.Traces = append(rt.Traces, entry)
	}
}

// RecordHTTP records an HTTP call
func (rt *RequestTracer) RecordHTTP(method, url string, duration time.Duration, status int, err error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	details := map[string]interface{}{
		"method": method,
		"url":    url,
		"status": status,
	}

	entry := TraceEntry{
		Name:      fmt.Sprintf("HTTP %s", method),
		Type:      "http",
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Duration:  duration,
		Status:    "completed",
		Details:   details,
	}

	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
	}

	if len(rt.currentPath) > 0 {
		parent := rt.getCurrentTrace()
		if parent != nil {
			parent.Children = append(parent.Children, entry)
		}
	} else {
		rt.Traces = append(rt.Traces, entry)
	}
}

// RecordCustom records a custom event
func (rt *RequestTracer) RecordCustom(name string, details interface{}) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	entry := TraceEntry{
		Name:      name,
		Type:      "custom",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  0,
		Status:    "completed",
		Details:   details,
	}

	if len(rt.currentPath) > 0 {
		parent := rt.getCurrentTrace()
		if parent != nil {
			parent.Children = append(parent.Children, entry)
		}
	} else {
		rt.Traces = append(rt.Traces, entry)
	}
}

// GetStackTrace captures current goroutine stack
func (rt *RequestTracer) GetStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Complete marks the tracer as complete
func (rt *RequestTracer) Complete() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.EndTime = time.Now()

	// End any remaining open traces
	for len(rt.currentPath) > 0 {
		trace := rt.getCurrentTrace()
		if trace != nil && trace.Status == "running" {
			trace.EndTime = time.Now()
			trace.Duration = trace.EndTime.Sub(trace.StartTime)
			trace.Status = "completed"
		}
		rt.currentPath = rt.currentPath[:len(rt.currentPath)-1]
	}
}

// GetTraces returns all traces
func (rt *RequestTracer) GetTraces() []TraceEntry {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.Traces
}

// Helper methods

func (rt *RequestTracer) getCurrentTrace() *TraceEntry {
	if len(rt.currentPath) == 0 {
		return nil
	}

	// Check if root index is valid
	if rt.currentPath[0] >= len(rt.Traces) {
		return nil
	}

	trace := &rt.Traces[rt.currentPath[0]]
	for i := 1; i < len(rt.currentPath); i++ {
		if rt.currentPath[i] >= len(trace.Children) {
			return nil // Invalid path, cannot traverse further
		}
		trace = &trace.Children[rt.currentPath[i]]
	}
	return trace
}

func (rt *RequestTracer) getParentTrace() *TraceEntry {
	if len(rt.currentPath) <= 1 {
		return nil
	}

	// Check if root index is valid
	if rt.currentPath[0] >= len(rt.Traces) {
		return nil
	}

	trace := &rt.Traces[rt.currentPath[0]]
	for i := 1; i < len(rt.currentPath)-1; i++ {
		if rt.currentPath[i] >= len(trace.Children) {
			return nil // Invalid path, cannot traverse further
		}
		trace = &trace.Children[rt.currentPath[i]]
	}
	return trace
}

// Context key for request tracer
type tracerKey struct{}

// WithTracer adds a tracer to the context
func WithTracer(ctx context.Context, tracer *RequestTracer) context.Context {
	return context.WithValue(ctx, tracerKey{}, tracer)
}

// GetTracer gets the tracer from context
func GetTracer(ctx context.Context) *RequestTracer {
	if tracer, ok := ctx.Value(tracerKey{}).(*RequestTracer); ok {
		return tracer
	}
	return nil
}

// TraceMiddleware wraps a handler with automatic tracing
func TraceMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := GetTracer(r.Context())
			if tracer != nil {
				tracer.StartTrace(name, "middleware", nil)
				defer tracer.EndTrace(nil)
			}
			next.ServeHTTP(w, r)
		})
	}
}
