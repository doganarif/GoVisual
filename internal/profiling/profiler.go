package profiling

import (
	"bytes"
	"context"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

// ProfileType represents the type of profiling to perform
type ProfileType uint32

const (
	// ProfileNone disables all profiling
	ProfileNone ProfileType = 0
	// ProfileCPU enables CPU profiling
	ProfileCPU ProfileType = 1 << iota
	// ProfileMemory enables memory profiling
	ProfileMemory
	// ProfileGoroutine enables goroutine tracking
	ProfileGoroutine
	// ProfileBlocking enables blocking profiling
	ProfileBlocking
	// ProfileAll enables all profiling types
	ProfileAll = ProfileCPU | ProfileMemory | ProfileGoroutine | ProfileBlocking
)

// Metrics contains performance metrics for a request
type Metrics struct {
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
	SQLQueries       []SQLQueryMetric         `json:"sql_queries,omitempty"`
	HTTPCalls        []HTTPCallMetric         `json:"http_calls,omitempty"`
	Bottlenecks      []Bottleneck             `json:"bottlenecks,omitempty"`
	CPUProfile       []byte                   `json:"-"` // Raw CPU profile data
	HeapProfile      []byte                   `json:"-"` // Raw heap profile data
}

// SQLQueryMetric represents metrics for a SQL query
type SQLQueryMetric struct {
	Query    string        `json:"query"`
	Duration time.Duration `json:"duration"`
	Rows     int           `json:"rows"`
	Error    string        `json:"error,omitempty"`
}

// HTTPCallMetric represents metrics for an HTTP call
type HTTPCallMetric struct {
	Method   string        `json:"method"`
	URL      string        `json:"url"`
	Duration time.Duration `json:"duration"`
	Status   int           `json:"status"`
	Size     int64         `json:"size"`
}

// Bottleneck represents a performance bottleneck
type Bottleneck struct {
	Type        string        `json:"type"` // "cpu", "memory", "io", "database", "http"
	Description string        `json:"description"`
	Impact      float64       `json:"impact"` // 0-1 scale of impact
	Duration    time.Duration `json:"duration"`
	Suggestion  string        `json:"suggestion"`
}

// Profiler handles performance profiling for requests
type Profiler struct {
	enabled          atomic.Bool
	profileType      atomic.Uint32
	threshold        time.Duration // Minimum duration to trigger profiling
	mu               sync.RWMutex
	metrics          map[string]*Metrics
	maxMetrics       int
	cpuProfileMu     sync.Mutex         // Protects CPU profiling global state
	activeCPUProfile *cpuProfileSession // Currently active CPU profile session
}

// cpuProfileSession represents an active CPU profiling session
type cpuProfileSession struct {
	requestID string
	buffer    bytes.Buffer
	started   bool
}

// NewProfiler creates a new profiler instance
func NewProfiler(maxMetrics int) *Profiler {
	if maxMetrics <= 0 {
		maxMetrics = 1000
	}
	p := &Profiler{
		threshold:  10 * time.Millisecond, // Default threshold
		metrics:    make(map[string]*Metrics),
		maxMetrics: maxMetrics,
	}
	p.enabled.Store(true)
	p.profileType.Store(uint32(ProfileAll))
	return p
}

// SetEnabled enables or disables profiling
func (p *Profiler) SetEnabled(enabled bool) {
	p.enabled.Store(enabled)
}

// SetProfileType sets the types of profiling to perform
func (p *Profiler) SetProfileType(pt ProfileType) {
	p.profileType.Store(uint32(pt))
}

// SetThreshold sets the minimum duration to trigger profiling
func (p *Profiler) SetThreshold(threshold time.Duration) {
	p.mu.Lock()
	p.threshold = threshold
	p.mu.Unlock()
}

// StartProfiling starts profiling for a request
func (p *Profiler) StartProfiling(ctx context.Context, requestID string) context.Context {
	if !p.enabled.Load() {
		return ctx
	}

	metrics := &Metrics{
		RequestID:       requestID,
		StartTime:       time.Now(),
		FunctionTimings: make(map[string]time.Duration),
		SQLQueries:      make([]SQLQueryMetric, 0),
		HTTPCalls:       make([]HTTPCallMetric, 0),
		Bottlenecks:     make([]Bottleneck, 0),
	}

	// Start CPU profiling if enabled (thread-safe)
	if p.hasProfile(ProfileCPU) {
		p.cpuProfileMu.Lock()
		if p.activeCPUProfile == nil {
			// No active CPU profile, start one for this request
			session := &cpuProfileSession{
				requestID: requestID,
				buffer:    bytes.Buffer{},
				started:   false,
			}
			if err := pprof.StartCPUProfile(&session.buffer); err == nil {
				session.started = true
				p.activeCPUProfile = session
			}
		}
		p.cpuProfileMu.Unlock()
	}

	// Capture initial memory stats
	if p.hasProfile(ProfileMemory) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		metrics.MemoryAlloc = m.Alloc
		metrics.NumGC = m.NumGC
		metrics.GCPauseTotal = time.Duration(m.PauseTotalNs)
	}

	// Capture initial goroutine count
	if p.hasProfile(ProfileGoroutine) {
		metrics.NumGoroutines = runtime.NumGoroutine()
	}

	// Store metrics in context
	ctx = context.WithValue(ctx, profileContextKey{}, metrics)

	// Store in profiler
	p.storeMetrics(requestID, metrics)

	return ctx
}

// EndProfiling ends profiling for a request
func (p *Profiler) EndProfiling(ctx context.Context) *Metrics {
	if !p.enabled.Load() {
		return nil
	}

	metrics, ok := ctx.Value(profileContextKey{}).(*Metrics)
	if !ok || metrics == nil {
		return nil
	}

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	// Skip profiling if below threshold
	p.mu.RLock()
	threshold := p.threshold
	p.mu.RUnlock()

	if metrics.Duration < threshold {
		// Stop CPU profiling if this request started it
		p.stopCPUProfilingIfActive(metrics.RequestID)
		p.removeMetrics(metrics.RequestID)
		return nil
	}

	// Stop CPU profiling and capture data if this request started it
	if p.hasProfile(ProfileCPU) {
		p.cpuProfileMu.Lock()
		if p.activeCPUProfile != nil && p.activeCPUProfile.requestID == metrics.RequestID && p.activeCPUProfile.started {
			pprof.StopCPUProfile()
			metrics.CPUProfile = p.activeCPUProfile.buffer.Bytes()
			p.activeCPUProfile = nil
		}
		p.cpuProfileMu.Unlock()
	}

	// Capture final memory stats
	if p.hasProfile(ProfileMemory) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		metrics.MemoryTotalAlloc = m.Alloc - metrics.MemoryAlloc
		metrics.MemoryAlloc = m.Alloc

		// Calculate GC impact
		gcPauseIncrease := time.Duration(m.PauseTotalNs) - metrics.GCPauseTotal
		if gcPauseIncrease > metrics.Duration/10 { // GC took more than 10% of time
			metrics.Bottlenecks = append(metrics.Bottlenecks, Bottleneck{
				Type:        "gc",
				Description: "High garbage collection overhead",
				Impact:      float64(gcPauseIncrease) / float64(metrics.Duration),
				Duration:    gcPauseIncrease,
				Suggestion:  "Consider optimizing memory allocations or increasing GOGC",
			})
		}
	}

	// Analyze bottlenecks
	p.analyzeBottlenecks(metrics)

	return metrics
}

// RecordFunction records the timing of a function
func (p *Profiler) RecordFunction(ctx context.Context, name string, fn func() error) error {
	if !p.enabled.Load() {
		return fn()
	}

	metrics, ok := ctx.Value(profileContextKey{}).(*Metrics)
	if !ok || metrics == nil {
		return fn()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	metrics.FunctionTimings[name] = duration

	return err
}

// RecordSQLQuery records metrics for a SQL query
func (p *Profiler) RecordSQLQuery(ctx context.Context, query string, duration time.Duration, rows int, err error) {
	if !p.enabled.Load() {
		return
	}

	metrics, ok := ctx.Value(profileContextKey{}).(*Metrics)
	if !ok || metrics == nil {
		return
	}

	metric := SQLQueryMetric{
		Query:    query,
		Duration: duration,
		Rows:     rows,
	}
	if err != nil {
		metric.Error = err.Error()
	}

	metrics.SQLQueries = append(metrics.SQLQueries, metric)

	// Also record in tracer if available
	if tracer, ok := ctx.Value(tracerKey{}).(interface {
		RecordSQL(string, time.Duration, int, error)
	}); ok {
		tracer.RecordSQL(query, duration, rows, err)
	}
}

// RecordHTTPCall records metrics for an HTTP call
func (p *Profiler) RecordHTTPCall(ctx context.Context, method, url string, duration time.Duration, status int, size int64) {
	if !p.enabled.Load() {
		return
	}

	metrics, ok := ctx.Value(profileContextKey{}).(*Metrics)
	if !ok || metrics == nil {
		return
	}

	metrics.HTTPCalls = append(metrics.HTTPCalls, HTTPCallMetric{
		Method:   method,
		URL:      url,
		Duration: duration,
		Status:   status,
		Size:     size,
	})

	// Also record in tracer if available
	if tracer, ok := ctx.Value(tracerKey{}).(interface {
		RecordHTTP(string, string, time.Duration, int, error)
	}); ok {
		tracer.RecordHTTP(method, url, duration, status, nil)
	}
}

// GetMetrics retrieves metrics for a request
func (p *Profiler) GetMetrics(requestID string) (*Metrics, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	metrics, ok := p.metrics[requestID]
	return metrics, ok
}

// GetAllMetrics retrieves all stored metrics
func (p *Profiler) GetAllMetrics() []*Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Metrics, 0, len(p.metrics))
	for _, m := range p.metrics {
		result = append(result, m)
	}
	return result
}

// ClearMetrics clears all stored metrics
func (p *Profiler) ClearMetrics() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics = make(map[string]*Metrics)
}

// analyzeBottlenecks analyzes metrics to identify bottlenecks
func (p *Profiler) analyzeBottlenecks(metrics *Metrics) {
	// Analyze SQL queries
	var totalSQLTime time.Duration
	slowestQuery := time.Duration(0)
	for _, q := range metrics.SQLQueries {
		totalSQLTime += q.Duration
		if q.Duration > slowestQuery {
			slowestQuery = q.Duration
		}
	}

	if totalSQLTime > metrics.Duration/2 {
		metrics.Bottlenecks = append(metrics.Bottlenecks, Bottleneck{
			Type:        "database",
			Description: "High database query time",
			Impact:      float64(totalSQLTime) / float64(metrics.Duration),
			Duration:    totalSQLTime,
			Suggestion:  "Consider optimizing queries, adding indexes, or implementing caching",
		})
	}

	// Analyze HTTP calls
	var totalHTTPTime time.Duration
	for _, h := range metrics.HTTPCalls {
		totalHTTPTime += h.Duration
	}

	if totalHTTPTime > metrics.Duration/3 {
		metrics.Bottlenecks = append(metrics.Bottlenecks, Bottleneck{
			Type:        "http",
			Description: "High external HTTP call time",
			Impact:      float64(totalHTTPTime) / float64(metrics.Duration),
			Duration:    totalHTTPTime,
			Suggestion:  "Consider implementing caching, batching requests, or using connection pooling",
		})
	}

	// Analyze memory allocations
	if metrics.MemoryTotalAlloc > 10*1024*1024 { // More than 10MB allocated
		metrics.Bottlenecks = append(metrics.Bottlenecks, Bottleneck{
			Type:        "memory",
			Description: "High memory allocation",
			Impact:      float64(metrics.MemoryTotalAlloc) / (10 * 1024 * 1024),
			Duration:    0,
			Suggestion:  "Consider reusing objects, using sync.Pool, or optimizing data structures",
		})
	}

	// Analyze function timings
	for name, duration := range metrics.FunctionTimings {
		if duration > metrics.Duration/4 {
			metrics.Bottlenecks = append(metrics.Bottlenecks, Bottleneck{
				Type:        "cpu",
				Description: "Slow function: " + name,
				Impact:      float64(duration) / float64(metrics.Duration),
				Duration:    duration,
				Suggestion:  "Consider optimizing algorithm or using concurrent processing",
			})
		}
	}
}

// Helper methods

func (p *Profiler) hasProfile(pt ProfileType) bool {
	return ProfileType(p.profileType.Load())&pt != 0
}

func (p *Profiler) storeMetrics(requestID string, metrics *Metrics) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Enforce max metrics limit
	if len(p.metrics) >= p.maxMetrics {
		// Remove oldest entry (simple FIFO for now)
		for k := range p.metrics {
			delete(p.metrics, k)
			break
		}
	}

	p.metrics[requestID] = metrics
}

func (p *Profiler) removeMetrics(requestID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.metrics, requestID)
}

type profileContextKey struct{}
type tracerKey struct{}

// stopCPUProfilingIfActive stops CPU profiling if the given request started it
func (p *Profiler) stopCPUProfilingIfActive(requestID string) {
	p.cpuProfileMu.Lock()
	defer p.cpuProfileMu.Unlock()

	if p.activeCPUProfile != nil && p.activeCPUProfile.requestID == requestID && p.activeCPUProfile.started {
		pprof.StopCPUProfile()
		p.activeCPUProfile = nil
	}
}

// SetTracer associates a tracer with the profiling context
func (p *Profiler) SetTracer(ctx context.Context, tracer interface{}) {
	// Tracer is already in context, no action needed
	// This method exists for compatibility
}

// ProfileWriter captures CPU profiles
type ProfileWriter struct {
	buf []byte
}

func (pw *ProfileWriter) Write(p []byte) (n int, err error) {
	pw.buf = append(pw.buf, p...)
	return len(p), nil
}
