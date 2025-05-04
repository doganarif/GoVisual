package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/doganarif/govisual"
)

// User represents a simple user model for JSON responses
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// TimingInfo represents detailed timing information for a request
type TimingInfo struct {
	StartTime      time.Time          `json:"-"`
	EndTime        time.Time          `json:"-"`
	Duration       time.Duration      `json:"duration_ms"`
	ConnectTime    time.Duration      `json:"connect_time_ms,omitempty"`
	DNSTime        time.Duration      `json:"dns_time_ms,omitempty"`
	TLSTime        time.Duration      `json:"tls_time_ms,omitempty"`
	FirstByteTime  time.Duration      `json:"ttfb_ms"`
	ProcessingTime time.Duration      `json:"processing_time_ms"`
	NetworkTime    time.Duration      `json:"network_time_ms"`
	Middleware     []MiddlewareTiming `json:"middleware_timings,omitempty"`
}

// MiddlewareTiming represents timing for a specific middleware
type MiddlewareTiming struct {
	Name      string        `json:"name"`
	StartTime time.Time     `json:"-"`
	EndTime   time.Time     `json:"-"`
	Duration  time.Duration `json:"duration_ms"`
}

// TraceResponseWriter is a custom response writer that captures response status and headers
type TraceResponseWriter struct {
	http.ResponseWriter
	status        int
	headerWritten bool
	body          *bytes.Buffer
}

// NewTraceResponseWriter creates a new TraceResponseWriter
func NewTraceResponseWriter(w http.ResponseWriter) *TraceResponseWriter {
	return &TraceResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // Default status
		headerWritten:  false,
		body:           new(bytes.Buffer),
	}
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (tw *TraceResponseWriter) WriteHeader(status int) {
	tw.status = status
	tw.headerWritten = true
	tw.ResponseWriter.WriteHeader(status)
}

// Write captures the response body and writes it to the underlying ResponseWriter
func (tw *TraceResponseWriter) Write(b []byte) (int, error) {
	if !tw.headerWritten {
		tw.headerWritten = true
		tw.ResponseWriter.WriteHeader(tw.status)
	}
	tw.body.Write(b)
	return tw.ResponseWriter.Write(b)
}

// Status returns the captured status code
func (tw *TraceResponseWriter) Status() int {
	return tw.status
}

// Body returns the captured response body
func (tw *TraceResponseWriter) Body() []byte {
	return tw.body.Bytes()
}

// TimingContextKey is a type for context keys related to timing
type TimingContextKey string

const (
	// ContextKeyTiming is the context key for timing information
	ContextKeyTiming TimingContextKey = "timing"
	// ContextKeyMiddleware is the context key for middleware information
	ContextKeyMiddleware TimingContextKey = "middleware"
	// ContextKeyRoute is the context key for route information
	ContextKeyRoute TimingContextKey = "route"

	// GoVisual built-in middleware trace format
	BuiltInTraceFormat = true // Set to true to use GoVisual's built-in trace format
)

func main() {
	// Create a simple mux
	mux := http.NewServeMux()

	// Wrap everything with timing middleware
	timedMux := withGlobalTiming(mux)

	// Root and dashboard info
	timedMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GoVisual Test Server</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1 { color: #3498db; }
        h2 { color: #2c3e50; margin-top: 30px; }
        pre { background-color: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .endpoint { margin-bottom: 10px; padding: 10px; border-left: 4px solid #3498db; background-color: #f8f9fa; }
        .method { font-weight: bold; color: #3498db; }
        .desc { margin-top: 5px; color: #7f8c8d; }
    </style>
</head>
<body>
    <h1>GoVisual Test Server</h1>
    <p>This is a test server to demonstrate the GoVisual dashboard.</p>
    <p><a href="/__viz" target="_blank">Click here to open the GoVisual dashboard</a></p>
    
    <h2>Available API Endpoints</h2>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/users</div>
        <div class="desc">Returns a list of users (fast response)</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/users/:id</div>
        <div class="desc">Returns a single user by ID</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">POST</span> /api/users</div>
        <div class="desc">Creates a new user (expects JSON body)</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">PUT</span> /api/users/:id</div>
        <div class="desc">Updates a user (expects JSON body)</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">DELETE</span> /api/users/:id</div>
        <div class="desc">Deletes a user</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/slow</div>
        <div class="desc">A slow endpoint (500ms response time)</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/very-slow</div>
        <div class="desc">A very slow endpoint (2s response time)</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/error</div>
        <div class="desc">Returns a 500 server error</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/not-found</div>
        <div class="desc">Returns a 404 not found error</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/unauthorized</div>
        <div class="desc">Returns a 401 unauthorized error</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/redirect</div>
        <div class="desc">Redirects to the home page</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/large-response</div>
        <div class="desc">Returns a large JSON response</div>
    </div>
    
    <h3>Middleware & Tracing Examples</h3>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/middleware/simple</div>
        <div class="desc">Request that passes through basic middleware</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/middleware/chain</div>
        <div class="desc">Request that passes through multiple middleware</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/middleware/slow</div>
        <div class="desc">Request with slow middleware processing</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/middleware/error</div>
        <div class="desc">Request where middleware returns an error</div>
    </div>
    
    <h3>Detailed Timing Examples</h3>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/timing/basic</div>
        <div class="desc">Shows basic request timing information</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/timing/detailed</div>
        <div class="desc">Shows detailed timing breakdown</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/timing/network-simulation</div>
        <div class="desc">Simulates network latency and processing time</div>
    </div>
    
    <h3>Route Matching Examples</h3>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/routes/users/:id</div>
        <div class="desc">Dynamic route with ID parameter</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/routes/products/:category/:id</div>
        <div class="desc">Nested dynamic route with multiple parameters</div>
    </div>
    
    <div class="endpoint">
        <div><span class="method">GET</span> /api/routes/regex/items/([0-9]+)</div>
        <div class="desc">Route with regex pattern matching</div>
    </div>
    
    <h2>Test the Dashboard</h2>
    <p>Make requests to these endpoints to see them appear in the GoVisual dashboard.</p>
    <p>Run <code>./test-dashboard.sh</code> to automatically test all endpoints.</p>
</body>
</html>
		`)
	})

	// GET /api/users - Get all users
	timedMux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			users := []User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now().Add(-24 * time.Hour)},
				{ID: 2, Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now().Add(-48 * time.Hour)},
				{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", CreatedAt: time.Now().Add(-72 * time.Hour)},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
			return
		} else if r.Method == "POST" {
			// Create a new user
			var user User
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid request body",
					Error:   err.Error(),
				})
				return
			}

			// Simulate created user
			user.ID = rand.Intn(1000) + 100
			user.CreatedAt = time.Now()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
			return
		}

		// Method not allowed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
	})

	// GET /api/users/:id - Get a specific user
	timedMux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		idStr := r.URL.Path[len("/api/users/"):]
		if idStr == "" {
			http.NotFound(w, r)
			return
		}

		// Add route trace information
		routeTraceData := map[string]interface{}{
			"pattern": "/api/users/:id",
			"path":    r.URL.Path,
			"params": map[string]string{
				"id": idStr,
			},
		}

		routeTraceJSON, _ := json.Marshal(routeTraceData)
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyRoute, string(routeTraceJSON)))

		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid user ID",
				Error:   err.Error(),
			})
			return
		}

		if r.Method == "GET" {
			// Get user by ID
			user := User{
				ID:        id,
				Name:      "User " + idStr,
				Email:     "user" + idStr + "@example.com",
				CreatedAt: time.Now().Add(-time.Duration(id) * 24 * time.Hour),
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		} else if r.Method == "PUT" {
			// Update user
			var updatedUser User
			err := json.NewDecoder(r.Body).Decode(&updatedUser)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{
					Status:  http.StatusBadRequest,
					Message: "Invalid request body",
					Error:   err.Error(),
				})
				return
			}

			// Simulate updated user
			updatedUser.ID = id

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedUser)
			return
		} else if r.Method == "DELETE" {
			// Delete user
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Method not allowed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusMethodNotAllowed,
			Message: "Method not allowed",
		})
	})

	// Slow endpoints
	timedMux.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "This was a slow request (500ms)",
			"duration": 500,
		})
	})

	timedMux.HandleFunc("/api/very-slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "This was a very slow request (2000ms)",
			"duration": 2000,
		})
	})

	// Error endpoints
	timedMux.HandleFunc("/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "An internal server error occurred",
			Error:   "Simulated server error for testing",
		})
	})

	timedMux.HandleFunc("/api/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "Resource not found",
		})
	})

	timedMux.HandleFunc("/api/unauthorized", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized access",
		})
	})

	// Redirect endpoint
	timedMux.HandleFunc("/api/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	// Large response
	timedMux.HandleFunc("/api/large-response", func(w http.ResponseWriter, r *http.Request) {
		// Generate a large array of items
		items := make([]map[string]interface{}, 100)
		for i := 0; i < 100; i++ {
			items[i] = map[string]interface{}{
				"id":          i + 1,
				"name":        fmt.Sprintf("Item %d", i+1),
				"description": fmt.Sprintf("This is item number %d with a longer description to increase payload size", i+1),
				"created_at":  time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
				"updated_at":  time.Now().Format(time.RFC3339),
				"metadata": map[string]interface{}{
					"category":    fmt.Sprintf("Category %d", (i%5)+1),
					"tags":        []string{"tag1", "tag2", "tag3"},
					"views":       rand.Intn(1000),
					"is_active":   rand.Intn(2) == 1,
					"coordinates": []float64{rand.Float64() * 100, rand.Float64() * 100},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total":  len(items),
			"items":  items,
			"page":   1,
			"limit":  100,
			"status": "success",
		})
	})

	// Middleware examples

	// Simple middleware
	timedMux.HandleFunc("/api/middleware/simple", withTracedMiddleware("auth", 10,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Request processed through simple middleware",
			})
		}))

	// Chain of middleware
	timedMux.HandleFunc("/api/middleware/chain", withTracedMiddleware("auth", 10,
		withTracedMiddleware("logging", 20,
			withTracedMiddleware("rate-limit", 15,
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"message": "Request processed through middleware chain",
					})
				}))))

	// Slow middleware
	timedMux.HandleFunc("/api/middleware/slow", withTracedMiddleware("slow-middleware", 300,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Request processed through slow middleware",
			})
		}))

	// Error in middleware
	timedMux.HandleFunc("/api/middleware/error", withErrorMiddleware("error-middleware",
		func(w http.ResponseWriter, r *http.Request) {
			// This will never be reached because the middleware returns an error
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "This won't be reached",
			})
		}))

	// Timing examples

	// Basic timing
	timedMux.HandleFunc("/api/timing/basic", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Basic timing information",
		})
	})

	// Detailed timing
	timedMux.HandleFunc("/api/timing/detailed", func(w http.ResponseWriter, r *http.Request) {
		// Simulate different phases of request processing
		dnsTime := simulateOperation(5)
		connectTime := simulateOperation(10)
		tlsTime := simulateOperation(15)
		processingTime := simulateOperation(70)

		// Store timing information in context
		timing := TimingInfo{
			StartTime:      time.Now().Add(-dnsTime - connectTime - tlsTime - processingTime),
			Duration:       dnsTime + connectTime + tlsTime + processingTime,
			DNSTime:        dnsTime,
			ConnectTime:    connectTime,
			TLSTime:        tlsTime,
			FirstByteTime:  dnsTime + connectTime + tlsTime + (processingTime / 2),
			ProcessingTime: processingTime,
			NetworkTime:    dnsTime + connectTime + tlsTime,
		}

		// Add timing to context
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyTiming, timing))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Detailed timing information",
			"timing": map[string]interface{}{
				"duration_ms":        timing.Duration.Milliseconds(),
				"dns_time_ms":        timing.DNSTime.Milliseconds(),
				"connect_time_ms":    timing.ConnectTime.Milliseconds(),
				"tls_time_ms":        timing.TLSTime.Milliseconds(),
				"ttfb_ms":            timing.FirstByteTime.Milliseconds(),
				"processing_time_ms": timing.ProcessingTime.Milliseconds(),
				"network_time_ms":    timing.NetworkTime.Milliseconds(),
			},
		})
	})

	// Network simulation
	timedMux.HandleFunc("/api/timing/network-simulation", func(w http.ResponseWriter, r *http.Request) {
		// Simulate network latency (higher than usual)
		networkLatency := simulateOperation(150)

		// Simulate processing
		processingTime := simulateOperation(200)

		// Store timing information in context
		timing := TimingInfo{
			StartTime:      time.Now().Add(-networkLatency - processingTime),
			Duration:       networkLatency + processingTime,
			FirstByteTime:  networkLatency,
			ProcessingTime: processingTime,
			NetworkTime:    networkLatency,
		}

		// Add timing to context
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyTiming, timing))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Network simulation timing",
			"timing": map[string]interface{}{
				"duration_ms":        timing.Duration.Milliseconds(),
				"ttfb_ms":            timing.FirstByteTime.Milliseconds(),
				"processing_time_ms": timing.ProcessingTime.Milliseconds(),
				"network_time_ms":    timing.NetworkTime.Milliseconds(),
			},
		})
	})

	// Route matching examples

	// Dynamic route with ID parameter
	timedMux.HandleFunc("/api/routes/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if the path follows the pattern /api/routes/users/:id
		if !strings.HasPrefix(path, "/api/routes/users/") {
			http.NotFound(w, r)
			return
		}

		// Extract ID from path
		idStr := path[len("/api/routes/users/"):]
		if idStr == "" {
			http.NotFound(w, r)
			return
		}

		// Add route trace information
		routeTraceData := map[string]interface{}{
			"pattern": "/api/routes/users/:id",
			"path":    path,
			"params": map[string]string{
				"id": idStr,
			},
		}

		routeTraceJSON, _ := json.Marshal(routeTraceData)
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyRoute, string(routeTraceJSON)))

		// Parse ID
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid user ID",
				Error:   err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Dynamic route with ID parameter",
			"route_info": fmt.Sprintf("/api/routes/users/%d", id),
			"params": map[string]interface{}{
				"id": id,
			},
		})
	})

	// Nested dynamic route with multiple parameters
	timedMux.HandleFunc("/api/routes/products/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if the path follows the pattern /api/routes/products/:category/:id
		if !strings.HasPrefix(path, "/api/routes/products/") {
			http.NotFound(w, r)
			return
		}

		// Extract parameters from path
		params := strings.Split(path[len("/api/routes/products/"):], "/")
		if len(params) != 2 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid path format, expected /api/routes/products/:category/:id",
			})
			return
		}

		category := params[0]
		idStr := params[1]

		// Add route trace information
		routeTraceData := map[string]interface{}{
			"pattern": "/api/routes/products/:category/:id",
			"path":    path,
			"params": map[string]string{
				"category": category,
				"id":       idStr,
			},
		}

		routeTraceJSON, _ := json.Marshal(routeTraceData)
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyRoute, string(routeTraceJSON)))

		// Parse ID
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid product ID",
				Error:   err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Nested dynamic route with multiple parameters",
			"route_info": fmt.Sprintf("/api/routes/products/%s/%d", category, id),
			"params": map[string]interface{}{
				"category": category,
				"id":       id,
			},
		})
	})

	// Route with regex pattern matching
	timedMux.HandleFunc("/api/routes/regex/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Use regex to match pattern /api/routes/regex/items/([0-9]+)
		pattern := regexp.MustCompile(`^/api/routes/regex/items/([0-9]+)$`)
		matches := pattern.FindStringSubmatch(path)

		if matches == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid path format, expected /api/routes/regex/items/{numeric_id}",
			})
			return
		}

		// Extract ID from regex match
		idStr := matches[1]
		id, _ := strconv.Atoi(idStr)

		// Add route trace information
		routeTraceData := map[string]interface{}{
			"pattern": "^/api/routes/regex/items/([0-9]+)$",
			"path":    path,
			"params": map[string]string{
				"id": idStr,
			},
			"matches": matches,
		}

		routeTraceJSON, _ := json.Marshal(routeTraceData)
		r = r.WithContext(context.WithValue(r.Context(), ContextKeyRoute, string(routeTraceJSON)))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Route with regex pattern matching",
			"route_info": fmt.Sprintf("/api/routes/regex/items/%d", id),
			"params": map[string]interface{}{
				"id": id,
			},
			"regex_pattern": "^/api/routes/regex/items/([0-9]+)$",
		})
	})

	// Health check and monitoring endpoints
	timedMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	timedMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "# HELP api_requests_total The total number of API requests\n")
		fmt.Fprintf(w, "# TYPE api_requests_total counter\n")
		fmt.Fprintf(w, "api_requests_total %d\n", rand.Intn(1000))
		fmt.Fprintf(w, "# HELP api_request_duration_seconds The request duration in seconds\n")
		fmt.Fprintf(w, "# TYPE api_request_duration_seconds histogram\n")
		fmt.Fprintf(w, "api_request_duration_seconds_sum %f\n", rand.Float64()*100)
		fmt.Fprintf(w, "api_request_duration_seconds_count %d\n", rand.Intn(1000))
	})

	// Wrap with govisual
	handler := govisual.Wrap(timedMux,
		govisual.WithMaxRequests(200),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		// Ignore health checks, metrics endpoints, and all static assets
		govisual.WithIgnorePaths(
			"/health",
			"/metrics",
			"/static/*",      // Wildcard for all static files
			"/api/internal/", // Ignore all paths under this directory
			"/favicon.ico",
		),
	)

	// Start the server
	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println("Visit http://localhost:8080/__viz to see the dashboard")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Top-level middleware function to add tracing to all requests
func withGlobalTiming(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create TraceResponseWriter to capture response details
		trw := NewTraceResponseWriter(w)

		// Create new context with timing information
		ctx := context.WithValue(r.Context(), ContextKeyTiming, &TimingInfo{
			StartTime:  startTime,
			Middleware: []MiddlewareTiming{},
		})

		// Create middleware for tracing
		middlewareInfo := map[string]interface{}{
			"stack": []map[string]interface{}{
				{
					"name":       "global-timing",
					"start_time": startTime.UnixNano() / int64(time.Millisecond),
				},
			},
		}

		// Add middleware info to context
		ctx = context.WithValue(ctx, ContextKeyMiddleware, middlewareInfo)

		// Process the request
		handler.ServeHTTP(trw, r.WithContext(ctx))

		// Get the timing info from context
		timingValue := r.Context().Value(ContextKeyTiming)

		// Record end time
		endTime := time.Now()
		processingTime := endTime.Sub(startTime)

		// Set response headers for tracing
		w.Header().Set("X-Processing-Time", fmt.Sprintf("%d", processingTime.Milliseconds()))
		w.Header().Set("X-Response-Time", fmt.Sprintf("%d", processingTime.Milliseconds()))

		// Add middleware trace headers if we have middleware info
		middlewareValue := r.Context().Value(ContextKeyMiddleware)
		if middlewareInfo, ok := middlewareValue.(map[string]interface{}); ok {
			if stack, ok := middlewareInfo["stack"].([]map[string]interface{}); ok && len(stack) > 0 {
				middlewareJSON, _ := json.Marshal(stack)
				w.Header().Set("X-Middleware-Trace", string(middlewareJSON))
			}
		}

		// Add route trace headers if we have route info
		routeValue := r.Context().Value(ContextKeyRoute)
		if routeInfo, ok := routeValue.(string); ok {
			w.Header().Set("X-Route-Trace", routeInfo)
		}
	})
}

// Middleware for tracing with name and simulated delay
func withTracedMiddleware(name string, delayMs int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Simulate processing
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		// Get existing middleware info
		var middlewareInfo map[string]interface{}
		middlewareValue := r.Context().Value(ContextKeyMiddleware)
		if middlewareValue != nil {
			if mi, ok := middlewareValue.(map[string]interface{}); ok {
				middlewareInfo = mi
			}
		}

		if middlewareInfo == nil {
			middlewareInfo = map[string]interface{}{
				"stack": []map[string]interface{}{},
			}
		}

		// Get the stack
		var stack []map[string]interface{}
		if s, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
			stack = s
		} else {
			stack = []map[string]interface{}{}
		}

		// Add this middleware to the stack
		stack = append(stack, map[string]interface{}{
			"name":       name,
			"start_time": startTime.UnixNano() / int64(time.Millisecond),
			"end_time":   time.Now().UnixNano() / int64(time.Millisecond),
			"duration":   time.Since(startTime).Milliseconds(),
		})

		// Update the stack
		middlewareInfo["stack"] = stack

		// Update the context
		ctx := context.WithValue(r.Context(), ContextKeyMiddleware, middlewareInfo)

		// Add header to trace middleware
		for i, mw := range stack {
			w.Header().Set(fmt.Sprintf("X-Middleware-%d-Name", i), fmt.Sprintf("%v", mw["name"]))
			w.Header().Set(fmt.Sprintf("X-Middleware-%d-Duration", i), fmt.Sprintf("%v", mw["duration"]))
		}

		// Call the next handler
		next(w, r.WithContext(ctx))
	}
}

// ErrorMiddleware is a middleware that returns an error
func withErrorMiddleware(name string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Simulate some processing
		time.Sleep(50 * time.Millisecond)

		// Get existing middleware info
		var middlewareInfo map[string]interface{}
		middlewareValue := r.Context().Value(ContextKeyMiddleware)
		if middlewareValue != nil {
			if mi, ok := middlewareValue.(map[string]interface{}); ok {
				middlewareInfo = mi
			}
		}

		if middlewareInfo == nil {
			middlewareInfo = map[string]interface{}{
				"stack": []map[string]interface{}{},
			}
		}

		// Get the stack
		var stack []map[string]interface{}
		if s, ok := middlewareInfo["stack"].([]map[string]interface{}); ok {
			stack = s
		} else {
			stack = []map[string]interface{}{}
		}

		// Add this middleware to the stack with error flag
		stack = append(stack, map[string]interface{}{
			"name":          name,
			"start_time":    startTime.UnixNano() / int64(time.Millisecond),
			"end_time":      time.Now().UnixNano() / int64(time.Millisecond),
			"duration":      time.Since(startTime).Milliseconds(),
			"error":         true,
			"error_message": "Authentication failed",
		})

		// Update the stack
		middlewareInfo["stack"] = stack

		// Update the context
		ctx := context.WithValue(r.Context(), ContextKeyMiddleware, middlewareInfo)

		// Add header to trace middleware
		for i, mw := range stack {
			w.Header().Set(fmt.Sprintf("X-Middleware-%d-Name", i), fmt.Sprintf("%v", mw["name"]))
			w.Header().Set(fmt.Sprintf("X-Middleware-%d-Duration", i), fmt.Sprintf("%v", mw["duration"]))
			if errFlag, ok := mw["error"].(bool); ok && errFlag {
				w.Header().Set(fmt.Sprintf("X-Middleware-%d-Error", i), "true")
				if errMsg, ok := mw["error_message"].(string); ok {
					w.Header().Set(fmt.Sprintf("X-Middleware-%d-Error-Message", i), errMsg)
				}
			}
		}

		// Return an error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Status:  http.StatusForbidden,
			Message: "Access denied by middleware",
			Error:   "Authentication failed",
		})

		// Don't call the next handler
	}
}

// simulateOperation simulates an operation taking the specified time
func simulateOperation(ms int) time.Duration {
	duration := time.Duration(ms) * time.Millisecond
	time.Sleep(duration)
	return duration
}
