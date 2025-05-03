package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/store"
)

// PathMatcher defines an interface for checking if a path should be ignored
type PathMatcher interface {
	ShouldIgnorePath(path string) bool
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code and response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	buffer     *bytes.Buffer
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write captures the response body
func (w *responseWriter) Write(b []byte) (int, error) {
	// Write to the buffer
	if w.buffer != nil {
		w.buffer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// Wrap wraps an http.Handler with the request visualization middleware
func Wrap(handler http.Handler, store store.Store, logRequestBody, logResponseBody bool, pathMatcher PathMatcher) http.Handler {
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
			// Read the body
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body.Close()

			// Store the body in the log
			reqLog.RequestBody = string(bodyBytes)

			// Create a new body for the request
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create response writer wrapper
		var resWriter *responseWriter
		if logResponseBody {
			resWriter = &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
				buffer:         &bytes.Buffer{},
			}
		} else {
			resWriter = &responseWriter{
				ResponseWriter: w,
				statusCode:     200, // Default status code
			}
		}

		// Record start time
		start := time.Now()

		// Call the handler
		handler.ServeHTTP(resWriter, r)

		// Calculate duration
		duration := time.Since(start)
		reqLog.Duration = duration.Milliseconds()

		// Capture response info
		reqLog.StatusCode = resWriter.statusCode

		// Capture response body if enabled
		if logResponseBody && resWriter.buffer != nil {
			reqLog.ResponseBody = resWriter.buffer.String()
		}

		// Store the request log
		store.Add(reqLog)
	})
}
