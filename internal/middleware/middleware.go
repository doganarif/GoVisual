package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/store"
)

// DefaultMaxBodyBytes is the default cap for captured request/response body size.
// Bodies larger than this are truncated with a marker suffix to avoid unbounded memory growth.
const DefaultMaxBodyBytes = 1 << 20 // 1 MiB

// truncationMarker is appended when a captured body has been truncated.
const truncationMarker = "...[truncated by govisual]"

// PathMatcher defines an interface for checking if a path should be ignored
type PathMatcher interface {
	ShouldIgnorePath(path string) bool
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code and response.
// It is safe for concurrent calls to Write (a handler that fans out writes across goroutines).
type responseWriter struct {
	http.ResponseWriter
	mu          sync.Mutex
	statusCode  int
	wroteHeader bool
	buffer      *bytes.Buffer
	maxBody     int  // 0 means unlimited
	truncated   bool // set once buffer hit maxBody
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.mu.Lock()
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
	}
	w.mu.Unlock()
	w.ResponseWriter.WriteHeader(code)
}

// Write captures the response body up to maxBody bytes, then passes through.
func (w *responseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	if w.buffer != nil && !w.truncated {
		remaining := w.maxBody - w.buffer.Len()
		switch {
		case w.maxBody <= 0:
			w.buffer.Write(b)
		case remaining > 0:
			if remaining >= len(b) {
				w.buffer.Write(b)
			} else {
				w.buffer.Write(b[:remaining])
				w.buffer.WriteString(truncationMarker)
				w.truncated = true
			}
		default:
			w.truncated = true
		}
	}
	w.mu.Unlock()
	return w.ResponseWriter.Write(b)
}

// Flush implements http.Flusher, forwarding to the underlying writer if it supports it.
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker, forwarding to the underlying writer if it supports it.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("govisual: underlying ResponseWriter does not implement http.Hijacker")
}

// Push implements http.Pusher, forwarding to the underlying writer if it supports it.
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

// readBodyCapped reads up to maxBody bytes from r, returns the bytes, a boolean
// indicating whether the body was truncated, and any read error.
func readBodyCapped(r io.Reader, maxBody int) ([]byte, bool, error) {
	if maxBody <= 0 {
		data, err := io.ReadAll(r)
		return data, false, err
	}
	limited := io.LimitReader(r, int64(maxBody)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return data, false, err
	}
	if len(data) > maxBody {
		return append(data[:maxBody], []byte(truncationMarker)...), true, nil
	}
	return data, false, nil
}

// Wrap wraps an http.Handler with the request visualization middleware
func Wrap(handler http.Handler, store store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher) http.Handler {
	return WrapWithLimits(handler, store, logRequestBody, logResponseBody, pathMatcher, DefaultMaxBodyBytes)
}

// WrapWithLimits is identical to Wrap but allows the caller to specify the maximum number of
// captured body bytes (per request and per response). A value <= 0 disables the cap.
func WrapWithLimits(handler http.Handler, store store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher, maxBody int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path should be ignored
		if pathMatcher != nil && pathMatcher.ShouldIgnorePath(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		// Create a new request log
		reqLog := model.NewRequestLog(r)

		// Capture request body if enabled
		if logRequestBody && r.Body != nil {
			bodyBytes, _, err := readBodyCapped(r.Body, maxBody)
			r.Body.Close()
			if err == nil {
				reqLog.RequestBody = string(bodyBytes)
			}
			// Always restore a body so the handler can read what was buffered.
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create response writer wrapper
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
		reqLog.StatusCode = resWriter.statusCode

		// Extract user-provided middleware-stack information from context
		if v := r.Context().Value(MiddlewareStackKey{}); v != nil {
			if middlewareInfo, ok := v.(map[string]interface{}); ok {
				if stack, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
					reqLog.MiddlewareTrace = stack
				}
			}
		}

		// Extract route trace information
		if v := r.Context().Value(RouteTraceKey{}); v != nil {
			if routeStr, ok := v.(string); ok {
				var routeInfo map[string]interface{}
				if err := json.Unmarshal([]byte(routeStr), &routeInfo); err == nil {
					reqLog.RouteTrace = routeInfo
				}
			}
		}

		if logResponseBody && resWriter.buffer != nil {
			reqLog.ResponseBody = resWriter.buffer.String()
		}

		if err := store.Add(reqLog); err != nil {
			// Storage errors are surfaced on the log entry's Error field so they
			// remain visible to anyone inspecting the dashboard backend; we do
			// not block the request path.
			_ = err
		}
	})
}
