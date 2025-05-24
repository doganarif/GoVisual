package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/options"
	"github.com/doganarif/govisual/internal/store"
	"github.com/doganarif/govisual/internal/utils"
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
func Wrap(handler http.Handler, store store.Store, config *options.Config, pathMatcher PathMatcher) http.Handler {

	logger := utils.NewLogger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path should be ignored
		if pathMatcher != nil && pathMatcher.ShouldIgnorePath(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		// Create a new request log
		reqLog := model.NewRequestLog(r)

		// Capture request body if enabled
		if config.LogRequestBody && r.Body != nil {
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
		if config.LogResponseBody {
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

		// Extract middleware information from context
		if middlewareValue := r.Context().Value("middleware"); middlewareValue != nil {
			if middlewareInfo, ok := middlewareValue.(map[string]interface{}); ok {
				if stack, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
					reqLog.MiddlewareTrace = stack
				}
			}
		}

		// Extract route trace information
		if routeValue := r.Context().Value("route"); routeValue != nil {
			if routeStr, ok := routeValue.(string); ok {
				var routeInfo map[string]interface{}
				if err := json.Unmarshal([]byte(routeStr), &routeInfo); err == nil {
					reqLog.RouteTrace = routeInfo
				}
			}
		}

		// Capture response body if enabled
		if config.LogRequestBody && resWriter.buffer != nil {
			reqLog.ResponseBody = resWriter.buffer.String()
		}

		// Log to console if enabled
		if config.LogRequestToConsole {
			logger.LogRequest(reqLog)
		}

		// Store the request log
		store.Add(reqLog)
	})
}
