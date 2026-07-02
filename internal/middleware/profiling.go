package middleware

import (
	"bytes"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

// ProfilingConfig contains configuration for profiling middleware
type ProfilingConfig struct {
	Enabled       bool
	ProfileType   profiling.ProfileType
	Threshold     time.Duration
	CaptureTraces bool
}

// WrapWithProfiling wraps an http.Handler with request visualization and performance profiling.
func WrapWithProfiling(handler http.Handler, st store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher, profiler *profiling.Profiler) http.Handler {
	return WrapWithProfilingAndLimits(handler, st, logRequestBody, logResponseBody, pathMatcher, profiler, DefaultMaxBodyBytes, 1)
}

// WrapWithProfilingAndLimits is identical to WrapWithProfiling but exposes the captured-body size cap.
func WrapWithProfilingAndLimits(handler http.Handler, st store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher, profiler *profiling.Profiler, maxBody int, sampleRate float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if pathMatcher != nil && pathMatcher.ShouldIgnorePath(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		if sampleRate < 1 && rand.Float64() >= sampleRate {
			handler.ServeHTTP(w, r)
			return
		}

		reqLog := store.NewRequestLog(r)

		tracer := NewRequestTracer(reqLog.ID)
		tracer.StartTrace("Request Handler", "handler", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})

		ctx, collector := WithLogCollector(r.Context())
		ctx = WithTracer(ctx, tracer)
		// Register the tracer as a TracerSink so the profiler forwards SQL/HTTP
		// events into the tracer's child traces.
		ctx = profiling.WithTracerSink(ctx, tracer)

		if profiler != nil {
			ctx = profiler.StartProfiling(ctx, reqLog.ID)
		}
		r = r.WithContext(ctx)

		if logRequestBody && r.Body != nil {
			bodyBytes, _, err := readBodyCapped(r.Body, maxBody)
			r.Body.Close()
			if err == nil {
				reqLog.RequestBody = string(bodyBytes)
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			maxBody:        maxBody,
		}
		if logResponseBody {
			resWriter.buffer = &bytes.Buffer{}
		}

		start := time.Now()
		handler.ServeHTTP(resWriter, r)
		reqLog.Duration = time.Since(start).Milliseconds()

		if profiler != nil {
			if metrics := profiler.EndProfiling(ctx); metrics != nil {
				reqLog.PerformanceMetrics = metrics.Model()
			}
		}

		reqLog.StatusCode = resWriter.status()

		tracer.EndTrace(nil)
		tracer.Complete()

		reqLog.MiddlewareTrace = make([]map[string]interface{}, 0)
		for _, trace := range tracer.GetTraces() {
			traceMap := map[string]interface{}{
				"name":       trace.Name,
				"type":       trace.Type,
				"start_time": trace.StartTime,
				"end_time":   trace.EndTime,
				"duration":   trace.Duration.Milliseconds(),
				"status":     trace.Status,
				"details":    trace.Details,
				"children":   trace.Children,
			}
			if trace.Error != "" {
				traceMap["error"] = trace.Error
			}
			reqLog.MiddlewareTrace = append(reqLog.MiddlewareTrace, traceMap)
		}

		if v := r.Context().Value(MiddlewareStackKey{}); v != nil {
			if middlewareInfo, ok := v.(map[string]interface{}); ok {
				if stack, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
					reqLog.MiddlewareTrace = append(reqLog.MiddlewareTrace, stack...)
				}
			}
		}

		if logResponseBody {
			reqLog.ResponseBody = resWriter.body()
		}

		reqLog.Logs = collector.Snapshot()

		_ = st.Add(reqLog)
	})
}

// HTTPRoundTripper is a profiling HTTP round tripper for outgoing requests
type HTTPRoundTripper struct {
	Transport http.RoundTripper
	Profiler  *profiling.Profiler
}

// RoundTrip implements http.RoundTripper with profiling
func (rt *HTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if rt.Profiler == nil {
		return transport.RoundTrip(req)
	}

	start := time.Now()
	resp, err := transport.RoundTrip(req)
	duration := time.Since(start)

	if resp != nil {
		size := resp.ContentLength
		if size < 0 {
			size = 0
		}
		rt.Profiler.RecordHTTPCall(req.Context(), req.Method, req.URL.String(), duration, resp.StatusCode, size)
	} else {
		rt.Profiler.RecordHTTPCall(req.Context(), req.Method, req.URL.String(), duration, 0, 0)
	}
	return resp, err
}

// NewProfilingHTTPClient creates an HTTP client with profiling support
func NewProfilingHTTPClient(profiler *profiling.Profiler) *http.Client {
	return &http.Client{
		Transport: &HTTPRoundTripper{
			Transport: http.DefaultTransport,
			Profiler:  profiler,
		},
		Timeout: 30 * time.Second,
	}
}
