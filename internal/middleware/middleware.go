package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/doganarif/govisual/v2/store"
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

// status returns the captured status code. It takes the lock so it stays
// safe even if a handler goroutine is still writing.
func (w *responseWriter) status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.statusCode
}

// wrote reports whether the handler wrote a response header.
func (w *responseWriter) wrote() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.wroteHeader
}

// body returns a snapshot of the captured response body, or "" when body
// logging is off.
func (w *responseWriter) body() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer == nil {
		return ""
	}
	return w.buffer.String()
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
func Wrap(handler http.Handler, st store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher) http.Handler {
	return WrapWithLimits(handler, st, logRequestBody, logResponseBody, pathMatcher, DefaultMaxBodyBytes, 1)
}

// WrapWithLimits is identical to Wrap but allows the caller to specify the maximum number of
// captured body bytes (per request and per response, <= 0 disables the cap)
// and the sampling rate (0..1; requests that lose the coin toss pass through
// uncaptured).
func WrapWithLimits(handler http.Handler, st store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher, maxBody int, sampleRate float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path should be ignored
		if pathMatcher != nil && pathMatcher.ShouldIgnorePath(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		if sampleRate < 1 && rand.Float64() >= sampleRate {
			handler.ServeHTTP(w, r)
			return
		}

		// Create a new request log
		reqLog := store.NewRequestLog(r)

		// Collect slog output emitted with this request's context.
		ctx, collector := WithLogCollector(r.Context())
		r = r.WithContext(ctx)

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
		finish := func(panicked bool) {
			reqLog.Duration = time.Since(start).Milliseconds()
			reqLog.StatusCode = resWriter.status()
			if panicked && !resWriter.wrote() {
				// The handler died before writing; the client effectively
				// sees a failed request.
				reqLog.StatusCode = http.StatusInternalServerError
			}

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

			if logResponseBody {
				reqLog.ResponseBody = resWriter.body()
			}

			reqLog.Logs = collector.Snapshot()

			addToStore(st, reqLog)
		}

		defer func() {
			if rec := recover(); rec != nil {
				reqLog.Error = fmt.Sprintf("panic: %v", rec)
				reqLog.PanicStack = string(debug.Stack())
				finish(true)
				// Re-panic so recovery middleware and net/http behave exactly
				// as they would without govisual in the chain.
				panic(rec)
			}
		}()

		handler.ServeHTTP(resWriter, r)
		finish(false)
	})
}

// addToStore persists the entry. Storage errors are deliberately not allowed
// to block or fail the request path.
func addToStore(st store.Store, reqLog *store.RequestLog) {
	_ = st.Add(reqLog)
}
