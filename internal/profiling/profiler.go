package profiling

import (
	"bytes"
	"container/list"
	"context"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doganarif/govisual/v2/store"
)

// ProfileType represents the type of profiling to perform.
// Values are bitmask flags so multiple types can be OR'd together.
type ProfileType uint32

const (
	// ProfileNone disables all profiling
	ProfileNone ProfileType = 0
	// ProfileCPU enables CPU profiling
	ProfileCPU ProfileType = 1 << 0
	// ProfileMemory enables memory profiling
	ProfileMemory ProfileType = 1 << 1
	// ProfileGoroutine enables goroutine tracking
	ProfileGoroutine ProfileType = 1 << 2
	// ProfileBlocking enables blocking profiling
	ProfileBlocking ProfileType = 1 << 3
	// ProfileAll enables all profiling types
	ProfileAll = ProfileCPU | ProfileMemory | ProfileGoroutine | ProfileBlocking
)

// Metrics contains performance metrics for a request.
//
// Concurrency: a single request may fan out work across goroutines that each
// call RecordFunction/RecordSQLQuery/RecordHTTPCall on the same *Metrics. The
// mu field serializes those mutations. Readers (GetMetrics, JSON encoding,
// EndProfiling) snapshot under the same mutex.
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

	// Baseline of runtime.MemStats.TotalAlloc at StartProfiling, used to
	// compute MemoryTotalAlloc as a delta.
	totalAllocStart uint64

	mu sync.Mutex
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

// Profiler handles performance profiling for requests.
//
// Limitation: CPU profiling uses runtime/pprof.StartCPUProfile which is a
// process-global sampler. Only one request can be CPU-profiled at a time;
// concurrent requests that arrive while a profile is in progress will be
// captured for all other metrics (memory, goroutines, SQL, HTTP) but will
// not have a CPUProfile attached. This is a fundamental constraint of the
// Go runtime, not a bug.
type Profiler struct {
	enabled          atomic.Bool
	profileType      atomic.Uint32
	threshold        time.Duration // Minimum duration to trigger profiling
	mu               sync.RWMutex
	metrics          map[string]*list.Element // requestID -> *list.Element holding *Metrics
	order            *list.List               // FIFO insertion order of *Metrics
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
		metrics:    make(map[string]*list.Element),
		order:      list.New(),
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
		metrics.totalAllocStart = m.TotalAlloc
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

	metrics.mu.Lock()
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	duration := metrics.Duration
	metrics.mu.Unlock()

	// Skip profiling if below threshold
	p.mu.RLock()
	threshold := p.threshold
	p.mu.RUnlock()

	if duration < threshold {
		// Stop CPU profiling if this request started it
		p.stopCPUProfilingIfActive(metrics.RequestID)
		p.removeMetrics(metrics.RequestID)
		return nil
	}

	// Stop CPU profiling and capture data if this request started it
	var cpuProfile []byte
	if p.hasProfile(ProfileCPU) {
		p.cpuProfileMu.Lock()
		if p.activeCPUProfile != nil && p.activeCPUProfile.requestID == metrics.RequestID && p.activeCPUProfile.started {
			pprof.StopCPUProfile()
			cpuProfile = p.activeCPUProfile.buffer.Bytes()
			p.activeCPUProfile = nil
		}
		p.cpuProfileMu.Unlock()
	}

	// A leaked handler goroutine can still be calling Record*; every
	// mutation from here on stays under the metrics lock.
	metrics.mu.Lock()
	if cpuProfile != nil {
		metrics.CPUProfile = cpuProfile
	}

	// Capture final memory stats
	if p.hasProfile(ProfileMemory) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// TotalAlloc is monotonic, unlike Alloc, which shrinks after GC and
		// would underflow here.
		metrics.MemoryTotalAlloc = m.TotalAlloc - metrics.totalAllocStart
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
	metrics.mu.Unlock()

	return metrics.snapshot()
}

// snapshot returns a copy that is safe to hand outside the profiler while
// in-flight Record* calls may still mutate the original.
func (m *Metrics) snapshot() *Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := &Metrics{
		RequestID:        m.RequestID,
		StartTime:        m.StartTime,
		EndTime:          m.EndTime,
		Duration:         m.Duration,
		CPUTime:          m.CPUTime,
		MemoryAlloc:      m.MemoryAlloc,
		MemoryTotalAlloc: m.MemoryTotalAlloc,
		NumGoroutines:    m.NumGoroutines,
		NumGC:            m.NumGC,
		GCPauseTotal:     m.GCPauseTotal,
		CPUProfile:       m.CPUProfile,
		HeapProfile:      m.HeapProfile,
		totalAllocStart:  m.totalAllocStart,
	}
	if m.FunctionTimings != nil {
		c.FunctionTimings = make(map[string]time.Duration, len(m.FunctionTimings))
		for k, v := range m.FunctionTimings {
			c.FunctionTimings[k] = v
		}
	}
	c.SQLQueries = append([]SQLQueryMetric(nil), m.SQLQueries...)
	c.HTTPCalls = append([]HTTPCallMetric(nil), m.HTTPCalls...)
	c.Bottlenecks = append([]Bottleneck(nil), m.Bottlenecks...)
	return c
}

// Model converts the metrics into the plain store.PerformanceMetrics that
// gets attached to a RequestLog.
func (m *Metrics) Model() *store.PerformanceMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := &store.PerformanceMetrics{
		RequestID:        m.RequestID,
		StartTime:        m.StartTime,
		EndTime:          m.EndTime,
		Duration:         m.Duration,
		CPUTime:          m.CPUTime,
		MemoryAlloc:      m.MemoryAlloc,
		MemoryTotalAlloc: m.MemoryTotalAlloc,
		NumGoroutines:    m.NumGoroutines,
		NumGC:            m.NumGC,
		GCPauseTotal:     m.GCPauseTotal,
		CPUProfile:       m.CPUProfile,
		HeapProfile:      m.HeapProfile,
	}
	if m.FunctionTimings != nil {
		out.FunctionTimings = make(map[string]time.Duration, len(m.FunctionTimings))
		for k, v := range m.FunctionTimings {
			out.FunctionTimings[k] = v
		}
	}
	for _, q := range m.SQLQueries {
		out.SQLQueries = append(out.SQLQueries, store.SQLQuery(q))
	}
	for _, h := range m.HTTPCalls {
		out.HTTPCalls = append(out.HTTPCalls, store.HTTPCall(h))
	}
	for _, b := range m.Bottlenecks {
		out.Bottlenecks = append(out.Bottlenecks, store.Bottleneck(b))
	}
	return out
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

	metrics.mu.Lock()
	if metrics.FunctionTimings == nil {
		metrics.FunctionTimings = make(map[string]time.Duration)
	}
	metrics.FunctionTimings[name] = duration
	metrics.mu.Unlock()

	return err
}

// RecordSQLQuery records metrics for a SQL query
func (p *Profiler) RecordSQLQuery(ctx context.Context, query string, duration time.Duration, rows int, err error) {
	if !p.enabled.Load() {
		return
	}
	RecordSQL(ctx, query, duration, rows, err)
}

// RecordSQL attaches a query to the request profile carried by ctx. Without
// an active profile it is a no-op, so it is safe to call unconditionally —
// this is what the WrapDriver instrumentation uses.
func RecordSQL(ctx context.Context, query string, duration time.Duration, rows int, err error) {
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

	metrics.mu.Lock()
	metrics.SQLQueries = append(metrics.SQLQueries, metric)
	metrics.mu.Unlock()

	if sink := tracerSinkFromContext(ctx); sink != nil {
		sink.RecordSQL(query, duration, rows, err)
	}
}

// RecordHTTPCall records metrics for an HTTP call
func (p *Profiler) RecordHTTPCall(ctx context.Context, method, url string, duration time.Duration, status int, size int64) {
	if !p.enabled.Load() {
		return
	}
	RecordHTTP(ctx, method, url, duration, status, size)
}

// RecordHTTP attaches an outbound HTTP call to the request profile carried
// by ctx. Without an active profile it is a no-op, so it is safe to call
// unconditionally — this is what the WrapTransport instrumentation uses.
func RecordHTTP(ctx context.Context, method, url string, duration time.Duration, status int, size int64) {
	metrics, ok := ctx.Value(profileContextKey{}).(*Metrics)
	if !ok || metrics == nil {
		return
	}

	metrics.mu.Lock()
	metrics.HTTPCalls = append(metrics.HTTPCalls, HTTPCallMetric{
		Method:   method,
		URL:      url,
		Duration: duration,
		Status:   status,
		Size:     size,
	})
	metrics.mu.Unlock()

	if sink := tracerSinkFromContext(ctx); sink != nil {
		sink.RecordHTTP(method, url, duration, status, nil)
	}
}

// GetMetrics retrieves a snapshot of the metrics for a request
func (p *Profiler) GetMetrics(requestID string) (*Metrics, bool) {
	p.mu.RLock()
	elem, ok := p.metrics[requestID]
	p.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return elem.Value.(*Metrics).snapshot(), true
}

// GetAllMetrics retrieves snapshots of all stored metrics in FIFO insertion
// order (oldest first).
func (p *Profiler) GetAllMetrics() []*Metrics {
	p.mu.RLock()
	live := make([]*Metrics, 0, p.order.Len())
	for e := p.order.Front(); e != nil; e = e.Next() {
		live = append(live, e.Value.(*Metrics))
	}
	p.mu.RUnlock()

	result := make([]*Metrics, len(live))
	for i, m := range live {
		result[i] = m.snapshot()
	}
	return result
}

// ClearMetrics clears all stored metrics
func (p *Profiler) ClearMetrics() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics = make(map[string]*list.Element)
	p.order = list.New()
}

// analyzeBottlenecks analyzes metrics to identify bottlenecks.
// The caller must hold metrics.mu.
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

	// If the same requestID was already stored, replace its entry in place.
	if existing, ok := p.metrics[requestID]; ok {
		existing.Value = metrics
		return
	}

	// Real FIFO eviction: drop the front (oldest) element.
	for p.order.Len() >= p.maxMetrics {
		oldest := p.order.Front()
		if oldest == nil {
			break
		}
		oldMetrics := oldest.Value.(*Metrics)
		delete(p.metrics, oldMetrics.RequestID)
		p.order.Remove(oldest)
	}

	p.metrics[requestID] = p.order.PushBack(metrics)
}

func (p *Profiler) removeMetrics(requestID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if elem, ok := p.metrics[requestID]; ok {
		p.order.Remove(elem)
		delete(p.metrics, requestID)
	}
}

type profileContextKey struct{}

// TracerSink is the interface a tracer must implement to receive forwarded
// SQL and HTTP events recorded through the profiler. The middleware package's
// RequestTracer satisfies this interface.
type TracerSink interface {
	RecordSQL(query string, duration time.Duration, rows int, err error)
	RecordHTTP(method, url string, duration time.Duration, status int, err error)
}

type tracerSinkKey struct{}

// WithTracerSink attaches a TracerSink to the context so that calls to
// Profiler.RecordSQLQuery and Profiler.RecordHTTPCall are also forwarded
// to the tracer. Returns the new context.
func WithTracerSink(ctx context.Context, sink TracerSink) context.Context {
	if sink == nil {
		return ctx
	}
	return context.WithValue(ctx, tracerSinkKey{}, sink)
}

func tracerSinkFromContext(ctx context.Context) TracerSink {
	if v, ok := ctx.Value(tracerSinkKey{}).(TracerSink); ok {
		return v
	}
	return nil
}

// stopCPUProfilingIfActive stops CPU profiling if the given request started it
func (p *Profiler) stopCPUProfilingIfActive(requestID string) {
	p.cpuProfileMu.Lock()
	defer p.cpuProfileMu.Unlock()

	if p.activeCPUProfile != nil && p.activeCPUProfile.requestID == requestID && p.activeCPUProfile.started {
		pprof.StopCPUProfile()
		p.activeCPUProfile = nil
	}
}
