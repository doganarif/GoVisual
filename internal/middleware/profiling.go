package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/options"
	"github.com/doganarif/govisual/internal/profiling"
	"github.com/doganarif/govisual/internal/store"
)

// ProfilingConfig contains configuration for profiling middleware
type ProfilingConfig struct {
	Enabled       bool
	ProfileType   profiling.ProfileType
	Threshold     time.Duration
	CaptureTraces bool
}

// WrapWithProfiling wraps an http.Handler with request visualization and performance profiling
func WrapWithProfiling(handler http.Handler, store store.Store, config *options.Config, pathMatcher PathMatcher, profiler *profiling.Profiler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path should be ignored
		if pathMatcher != nil && pathMatcher.ShouldIgnorePath(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		// Create a new request log
		reqLog := model.NewRequestLog(r)

		// Create request tracer
		tracer := NewRequestTracer(reqLog.ID)
		tracer.StartTrace("Request Handler", "handler", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		})

		// Start profiling
		ctx := r.Context()
		ctx = WithTracer(ctx, tracer)

		if profiler != nil {
			ctx = profiler.StartProfiling(ctx, reqLog.ID)

			// Hook profiler to tracer
			profiler.SetTracer(ctx, tracer)
		}
		r = r.WithContext(ctx)

		// Capture request body if enabled
		if config.LogRequestBody && r.Body != nil {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body.Close()
			reqLog.RequestBody = string(bodyBytes)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create response writer wrapper with profiling support
		resWriter := &profilingResponseWriter{
			responseWriter: &responseWriter{
				ResponseWriter: w,
				statusCode:     200,
				buffer:         nil,
			},
			profiler: profiler,
			ctx:      ctx,
		}

		if config.LogResponseBody {
			resWriter.responseWriter.buffer = &bytes.Buffer{}
		}

		// Record start time
		start := time.Now()

		// Call the handler
		handler.ServeHTTP(resWriter, r)

		// Calculate duration
		duration := time.Since(start)
		reqLog.Duration = duration.Milliseconds()

		// End profiling and get metrics
		if profiler != nil {
			if metrics := profiler.EndProfiling(ctx); metrics != nil {
				reqLog.PerformanceMetrics = metrics
			}
		}

		// Capture response info
		reqLog.StatusCode = resWriter.responseWriter.statusCode

		// Complete the tracer
		tracer.EndTrace(nil)
		tracer.Complete()

		// Store traces in request log
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

		// Extract additional middleware information from context
		if middlewareValue := r.Context().Value("middleware"); middlewareValue != nil {
			if middlewareInfo, ok := middlewareValue.(map[string]interface{}); ok {
				if stack, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
					// Merge with existing traces
					for _, item := range stack {
						reqLog.MiddlewareTrace = append(reqLog.MiddlewareTrace, item)
					}
				}
			}
		}

		// Capture response body if enabled
		if config.LogResponseBody && resWriter.responseWriter.buffer != nil {
			reqLog.ResponseBody = resWriter.responseWriter.buffer.String()
		}

		// Store the request log
		store.Add(reqLog)
	})
}

// profilingResponseWriter extends responseWriter with profiling capabilities
type profilingResponseWriter struct {
	responseWriter *responseWriter
	profiler       *profiling.Profiler
	ctx            context.Context
}

func (w *profilingResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *profilingResponseWriter) WriteHeader(code int) {
	w.responseWriter.WriteHeader(code)
}

func (w *profilingResponseWriter) Write(b []byte) (int, error) {
	// Profile the write operation if significant
	if w.profiler != nil && len(b) > 1024 { // Only profile writes larger than 1KB
		return w.profileWrite(b)
	}
	return w.responseWriter.Write(b)
}

func (w *profilingResponseWriter) profileWrite(b []byte) (int, error) {
	var n int
	var err error

	w.profiler.RecordFunction(w.ctx, "response.Write", func() error {
		n, err = w.responseWriter.Write(b)
		return err
	})

	return n, err
}

// HTTPRoundTripper is a profiling HTTP round tripper for outgoing requests
type HTTPRoundTripper struct {
	Transport http.RoundTripper
	Profiler  *profiling.Profiler
}

// RoundTrip implements http.RoundTripper with profiling
func (rt *HTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.Profiler == nil {
		if rt.Transport != nil {
			return rt.Transport.RoundTrip(req)
		}
		return http.DefaultTransport.RoundTrip(req)
	}

	start := time.Now()

	transport := rt.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)

	duration := time.Since(start)

	// Record the HTTP call metrics
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
