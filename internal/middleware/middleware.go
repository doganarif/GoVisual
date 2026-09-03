package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	mu              sync.Mutex
	statusCode      int
	wroteHeader     bool
	responseHeaders http.Header
	buffer          *bytes.Buffer
	maxBody         int  // 0 means unlimited
	truncated       bool // set once buffer hit maxBody
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.ResponseWriter.WriteHeader(code)
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.responseHeaders = w.ResponseWriter.Header().Clone()
	}
}

// Write captures the response body up to maxBody bytes, then passes through.
func (w *responseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}

	written, err := w.ResponseWriter.Write(b)
	if w.responseHeaders == nil {
		w.responseHeaders = w.ResponseWriter.Header().Clone()
	}
	captured := written
	if captured < 0 {
		captured = 0
	}
	if captured > len(b) {
		captured = len(b)
	}
	if w.buffer != nil && !w.truncated && captured > 0 {
		remaining := w.maxBody - w.buffer.Len()
		switch {
		case w.maxBody <= 0:
			w.buffer.Write(b[:captured])
		case remaining > 0:
			if remaining >= captured {
				w.buffer.Write(b[:captured])
			} else {
				w.buffer.Write(b[:remaining])
				w.buffer.WriteString(truncationMarker)
				w.truncated = true
			}
		default:
			w.buffer.WriteString(truncationMarker)
			w.truncated = true
		}
	}
	return written, err
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

// headers returns the headers that were present when the response was
// committed. If the handler never committed a response, net/http will emit a
// 200 with the headers present when it returns, so snapshot those here.
func (w *responseWriter) headers() http.Header {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.responseHeaders != nil {
		return w.responseHeaders.Clone()
	}
	return w.ResponseWriter.Header().Clone()
}

// Unwrap exposes the underlying writer to http.ResponseController and other
// standard-library helpers without bypassing the capture wrapper.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) flushError() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	var err error
	switch underlying := w.ResponseWriter.(type) {
	case interface{ FlushError() error }:
		err = underlying.FlushError()
	case http.Flusher:
		underlying.Flush()
	default:
		return http.ErrNotSupported
	}
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
		w.responseHeaders = w.ResponseWriter.Header().Clone()
	}
	return err
}

func (w *responseWriter) hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("govisual: underlying ResponseWriter does not implement http.Hijacker")
}

func (w *responseWriter) push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

type responseFlusher struct{ writer *responseWriter }

func (f *responseFlusher) Flush() { _ = f.writer.flushError() }
func (f *responseFlusher) FlushError() error {
	return f.writer.flushError()
}

type responseHijacker struct{ writer *responseWriter }

func (h *responseHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.writer.hijack()
}

type responsePusher struct{ writer *responseWriter }

func (p *responsePusher) Push(target string, opts *http.PushOptions) error {
	return p.writer.push(target, opts)
}

// wrapResponseWriter exposes exactly the optional interfaces supported by the
// underlying writer. This keeps handler type assertions and ResponseController
// behavior identical to an unwrapped response.
func wrapResponseWriter(w *responseWriter) http.ResponseWriter {
	_, flushes := w.ResponseWriter.(http.Flusher)
	if _, ok := w.ResponseWriter.(interface{ FlushError() error }); ok {
		flushes = true
	}
	_, hijacks := w.ResponseWriter.(http.Hijacker)
	_, pushes := w.ResponseWriter.(http.Pusher)

	switch {
	case flushes && hijacks && pushes:
		return struct {
			*responseWriter
			*responseFlusher
			*responseHijacker
			*responsePusher
		}{w, &responseFlusher{w}, &responseHijacker{w}, &responsePusher{w}}
	case flushes && hijacks:
		return struct {
			*responseWriter
			*responseFlusher
			*responseHijacker
		}{w, &responseFlusher{w}, &responseHijacker{w}}
	case flushes && pushes:
		return struct {
			*responseWriter
			*responseFlusher
			*responsePusher
		}{w, &responseFlusher{w}, &responsePusher{w}}
	case hijacks && pushes:
		return struct {
			*responseWriter
			*responseHijacker
			*responsePusher
		}{w, &responseHijacker{w}, &responsePusher{w}}
	case flushes:
		return struct {
			*responseWriter
			*responseFlusher
		}{w, &responseFlusher{w}}
	case hijacks:
		return struct {
			*responseWriter
			*responseHijacker
		}{w, &responseHijacker{w}}
	case pushes:
		return struct {
			*responseWriter
			*responsePusher
		}{w, &responsePusher{w}}
	default:
		return w
	}
}

type replayReadCloser struct {
	prefix     *bytes.Reader
	body       io.ReadCloser
	pendingErr error
}

func (r *replayReadCloser) Read(p []byte) (int, error) {
	if r.prefix.Len() > 0 {
		n, err := r.prefix.Read(p)
		if n > 0 && r.prefix.Len() == 0 && r.pendingErr != nil {
			err = r.pendingErr
			r.pendingErr = nil
		}
		return n, err
	}
	if r.pendingErr != nil {
		err := r.pendingErr
		r.pendingErr = nil
		return 0, err
	}
	return r.body.Read(p)
}

func (r *replayReadCloser) Close() error {
	return r.body.Close()
}

// captureRequestBody reads only enough data to build the bounded capture and
// then reconstructs a stream containing every byte consumed followed by the
// untouched remainder. The handler therefore observes exactly the original
// request body even when the capture is truncated.
func captureRequestBody(body io.ReadCloser, maxBody int) ([]byte, io.ReadCloser, error) {
	if maxBody <= 0 {
		data, err := io.ReadAll(body)
		return data, &replayReadCloser{prefix: bytes.NewReader(data), body: body, pendingErr: err}, err
	}

	readLimit := int64(maxBody)
	if maxBody < int(^uint(0)>>1) {
		readLimit++
	}
	consumed, err := io.ReadAll(io.LimitReader(body, readLimit))
	restored := &replayReadCloser{
		prefix:     bytes.NewReader(consumed),
		body:       body,
		pendingErr: err,
	}
	if err != nil {
		return consumed, restored, err
	}
	if len(consumed) > maxBody {
		captured := make([]byte, 0, maxBody+len(truncationMarker))
		captured = append(captured, consumed[:maxBody]...)
		captured = append(captured, truncationMarker...)
		return captured, restored, nil
	}
	return consumed, restored, nil
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
	return WrapWithLimitsAndErrorHandler(handler, st, logRequestBody, logResponseBody, pathMatcher, maxBody, sampleRate, nil)
}

// WrapWithLimitsAndErrorHandler adds an optional callback for persistence
// failures. A nil callback logs the error and still leaves the request path
// unaffected.
func WrapWithLimitsAndErrorHandler(handler http.Handler, st store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher, maxBody int, sampleRate float64, onError func(error)) http.Handler {
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

		var requestBodyErr error
		// Capture request body if enabled
		if logRequestBody && r.Body != nil {
			bodyBytes, restoredBody, err := captureRequestBody(r.Body, maxBody)
			r.Body = restoredBody
			if err == nil {
				reqLog.RequestBody = string(bodyBytes)
			} else {
				requestBodyErr = fmt.Errorf("govisual: capture request body: %w", err)
			}
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
			reqLog.SetResponseHeaders(resWriter.headers())

			reqLog.Logs = collector.Snapshot()

			if requestBodyErr != nil {
				reportError(onError, requestBodyErr)
			}
			addToStore(st, reqLog, onError)
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

		handler.ServeHTTP(wrapResponseWriter(resWriter), r)
		finish(false)
	})
}

// addToStore persists the entry. Storage errors are deliberately not allowed
// to block or fail the request path.
func addToStore(st store.Store, reqLog *store.RequestLog, onError func(error)) {
	if err := st.Add(reqLog); err != nil {
		reportError(onError, fmt.Errorf("govisual: store request: %w", err))
	}
}

func reportError(onError func(error), err error) {
	if onError == nil {
		log.Printf("%v", err)
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("govisual: error handler panic: %v", recovered)
		}
	}()
	onError(err)
}
