package agent

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/transport"
)

// HTTPAgentConfig contains configuration options specific to HTTP agents.
type HTTPAgentConfig struct {
	AgentConfig

	// LogRequestBody determines whether request bodies are logged.
	LogRequestBody bool

	// LogResponseBody determines whether response bodies are logged.
	LogResponseBody bool

	// MaxBodySize is the maximum size of request/response bodies to log, in bytes.
	MaxBodySize int

	// IgnorePaths is a list of path patterns to ignore.
	IgnorePaths []string

	// IgnoreExtensions is a list of file extensions to ignore (e.g., ".jpg", ".png").
	IgnoreExtensions []string

	// PathTransformer is a function that transforms request paths before logging.
	PathTransformer func(string) string
}

// HTTPAgent is an agent that collects data from HTTP services.
type HTTPAgent struct {
	*BaseAgent
	config HTTPAgentConfig
}

// NewHTTPAgent creates a new HTTP agent with the given configuration.
func NewHTTPAgent(transportObj transport.Transport, opts ...HTTPOption) *HTTPAgent {
	config := HTTPAgentConfig{
		AgentConfig: AgentConfig{
			Transport: transportObj,
		},
		MaxBodySize: 1024 * 1024, // Default 1MB max body size
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	return &HTTPAgent{
		BaseAgent: NewBaseAgent("http", config.AgentConfig),
		config:    config,
	}
}

// HTTPOption is a function that configures an HTTP agent.
type HTTPOption func(*HTTPAgentConfig)

// WithHTTPRequestBodyLogging enables or disables logging of HTTP request bodies.
func WithHTTPRequestBodyLogging(enabled bool) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.LogRequestBody = enabled
	}
}

// WithHTTPResponseBodyLogging enables or disables logging of HTTP response bodies.
func WithHTTPResponseBodyLogging(enabled bool) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.LogResponseBody = enabled
	}
}

// WithMaxBodySize sets the maximum size of request/response bodies to log.
func WithMaxBodySize(size int) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.MaxBodySize = size
	}
}

// WithIgnorePaths sets the path patterns to ignore.
func WithIgnorePaths(patterns ...string) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// WithIgnoreExtensions sets the file extensions to ignore.
func WithIgnoreExtensions(extensions ...string) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.IgnoreExtensions = append(c.IgnoreExtensions, extensions...)
	}
}

// WithPathTransformer sets a function that transforms request paths before logging.
func WithPathTransformer(transformer func(string) string) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.PathTransformer = transformer
	}
}

// Middleware returns an HTTP middleware that captures request/response data.
func (a *HTTPAgent) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if path should be ignored
		if a.shouldIgnorePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Create a new request log
		reqLog := model.NewHTTPRequestLog(r)

		// Transform path if configured
		if a.config.PathTransformer != nil {
			reqLog.Path = a.config.PathTransformer(reqLog.Path)
		}

		// Capture request body if enabled
		if a.config.LogRequestBody && r.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, int64(a.config.MaxBodySize)))
			if err == nil {
				reqLog.RequestBody = string(bodyBytes)
				// Reset the body for downstream handlers
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Create response writer wrapper to capture response info
		rw := newResponseWriter(w, a.config.LogResponseBody, a.config.MaxBodySize)

		// Record start time
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record duration
		reqLog.Duration = time.Since(start).Milliseconds()

		// Capture response info
		reqLog.StatusCode = rw.Status()
		reqLog.ResponseHeaders = rw.Header()

		// Capture response body if enabled
		if a.config.LogResponseBody {
			reqLog.ResponseBody = rw.Body()
		}

		// Process the request log
		a.Process(reqLog)
	})
}

// shouldIgnorePath checks if a path should be ignored.
func (a *HTTPAgent) shouldIgnorePath(path string) bool {
	// Check if path is in the ignored paths list
	for _, pattern := range a.config.IgnorePaths {
		if pattern == path {
			return true
		}

		// Check for pattern matching
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}

	// Check if extension should be ignored
	for _, ext := range a.config.IgnoreExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code and response body.
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	buffer        *bytes.Buffer
	logBody       bool
	bodyWritten   bool
	maxBufferSize int
}

// newResponseWriter creates a new response writer wrapper.
func newResponseWriter(w http.ResponseWriter, logBody bool, maxSize int) *responseWriter {
	var buf *bytes.Buffer
	if logBody {
		buf = &bytes.Buffer{}
	}

	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
		buffer:         buf,
		logBody:        logBody,
		maxBufferSize:  maxSize,
	}
}

// WriteHeader captures the status code.
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures the response body.
func (w *responseWriter) Write(b []byte) (int, error) {
	// Capture body if enabled and not exceeding max size
	if w.logBody && w.buffer != nil && !w.bodyWritten && w.buffer.Len() < w.maxBufferSize {
		// Only write up to the max buffer size
		remaining := w.maxBufferSize - w.buffer.Len()
		if remaining <= 0 {
			w.bodyWritten = true
		} else if len(b) <= remaining {
			w.buffer.Write(b)
		} else {
			w.buffer.Write(b[:remaining])
			w.bodyWritten = true
		}
	}

	return w.ResponseWriter.Write(b)
}

// Status returns the captured status code.
func (w *responseWriter) Status() int {
	return w.statusCode
}

// Body returns the captured response body as a string.
func (w *responseWriter) Body() string {
	if w.buffer != nil {
		return w.buffer.String()
	}
	return ""
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

// Push implements the http.Pusher interface.
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return fmt.Errorf("underlying ResponseWriter does not implement http.Pusher")
}
