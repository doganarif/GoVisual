package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/doganarif/govisual"
)

// Define context keys
type contextKey string

const (
	// Use the exact same keys as in the govisual library
	middlewareContextKey contextKey = "middleware"
)

// Simple logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get existing middleware stack from context or create new one
		var middlewareInfo map[string]interface{}
		var stack []map[string]interface{}

		if existing := r.Context().Value(middlewareContextKey); existing != nil {
			if mi, ok := existing.(map[string]interface{}); ok {
				middlewareInfo = mi
				if s, ok := mi["stack"].([]map[string]interface{}); ok {
					stack = s
				}
			}
		}

		if middlewareInfo == nil {
			middlewareInfo = map[string]interface{}{
				"stack": []map[string]interface{}{},
			}
			stack = []map[string]interface{}{}
		}

		// Create this middleware's trace entry
		middlewareEntry := map[string]interface{}{
			"name":       "logging-middleware",
			"start_time": start.UnixMilli(),
		}

		// Add to stack
		stack = append(stack, middlewareEntry)
		middlewareInfo["stack"] = stack

		// Create new context with updated middleware info
		ctx := context.WithValue(r.Context(), middlewareContextKey, middlewareInfo)

		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))

		// Update the middleware entry with end time and duration
		endTime := time.Now()
		middlewareEntry["end_time"] = endTime.UnixMilli()
		middlewareEntry["duration"] = endTime.Sub(start).Milliseconds()

		// Log after request is processed
		fmt.Printf("[LOGGING] %s %s - %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}

// Response timing middleware
func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get existing middleware stack from context
		var middlewareInfo map[string]interface{}
		var stack []map[string]interface{}

		if existing := r.Context().Value(middlewareContextKey); existing != nil {
			if mi, ok := existing.(map[string]interface{}); ok {
				middlewareInfo = mi
				if s, ok := mi["stack"].([]map[string]interface{}); ok {
					stack = s
				}
			}
		}

		if middlewareInfo == nil {
			middlewareInfo = map[string]interface{}{
				"stack": []map[string]interface{}{},
			}
			stack = []map[string]interface{}{}
		}

		// Create this middleware's trace entry
		middlewareEntry := map[string]interface{}{
			"name":       "timing-middleware",
			"start_time": start.UnixMilli(),
		}

		// Add to stack
		stack = append(stack, middlewareEntry)
		middlewareInfo["stack"] = stack

		// Create new context with updated middleware info
		ctx := context.WithValue(r.Context(), middlewareContextKey, middlewareInfo)

		// Create a custom response writer to intercept the response
		responseWriter := &responseWriterWrapper{
			ResponseWriter: w,
			middlewareInfo: middlewareInfo,
		}

		next.ServeHTTP(responseWriter, r.WithContext(ctx))

		// Update the middleware entry with end time and duration
		endTime := time.Now()
		middlewareEntry["end_time"] = endTime.UnixMilli()
		middlewareEntry["duration"] = endTime.Sub(start).Milliseconds()

		// Add timing header
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", fmt.Sprintf("%v", duration.Milliseconds()))

		// Add trace header for GoVisual to pick up
		traceJSON, _ := json.Marshal(middlewareInfo)
		w.Header().Set("X-Middleware-Trace", string(traceJSON))
	})
}

// Custom response writer wrapper to capture response data
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode     int
	middlewareInfo map[string]interface{}
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode

	// Add trace header for GoVisual to pick up
	traceJSON, _ := json.Marshal(rw.middlewareInfo)
	rw.Header().Set("X-Middleware-Trace", string(traceJSON))

	rw.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	// Create a simple HTTP server
	mux := http.NewServeMux()

	// Add some example routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to my app!"))
	})

	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing time
		time.Sleep(200 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello, world!"}`))
	})

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		// Simulate database lookup
		time.Sleep(150 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id": 1, "name": "User 1"}, {"id": 2, "name": "User 2"}]`))
	})

	mux.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow endpoint
		time.Sleep(500 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "This was slow", "duration": 500}`))
	})

	mux.HandleFunc("/api/error", func(w http.ResponseWriter, r *http.Request) {
		// Simulate server error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	})

	// Apply middleware chain
	var handler http.Handler = mux
	handler = timingMiddleware(handler)
	handler = loggingMiddleware(handler)

	// Wrap with govisual
	visualHandler := govisual.Wrap(handler,
		govisual.WithMaxRequests(100),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("Visit http://localhost:8080/__viz to see the dashboard")
	fmt.Println("\nTry these endpoints:")
	fmt.Println("  http://localhost:8080/")
	fmt.Println("  http://localhost:8080/api/hello")
	fmt.Println("  http://localhost:8080/api/users")
	fmt.Println("  http://localhost:8080/api/slow")
	fmt.Println("  http://localhost:8080/api/error")

	log.Fatal(http.ListenAndServe(":8080", visualHandler))
}
