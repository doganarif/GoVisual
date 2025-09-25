package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/profiling"
	"github.com/doganarif/govisual/internal/store"
)

//go:embed static/*
var staticFiles embed.FS

// Handler is the HTTP handler for the dashboard
type Handler struct {
	store    store.Store
	profiler *profiling.Profiler
	staticFS fs.FS
}

// NewHandler creates a new dashboard handler
func NewHandler(store store.Store, profiler *profiling.Profiler) *Handler {
	// Create file system for static files
	staticFS, _ := fs.Sub(staticFiles, "static")

	return &Handler{
		store:    store,
		profiler: profiler,
		staticFS: staticFS,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API endpoints
	if strings.HasPrefix(r.URL.Path, "/api/") {
		switch r.URL.Path {
		case "/api/requests":
			h.handleAPIRequests(w, r)
		case "/api/events":
			h.handleSSE(w, r)
		case "/api/clear":
			h.handleClearRequests(w, r)
		case "/api/compare":
			h.handleCompareRequests(w, r)
		case "/api/replay":
			h.handleReplayRequest(w, r)
		case "/api/metrics":
			h.handleMetrics(w, r)
		case "/api/flamegraph":
			h.handleFlameGraph(w, r)
		case "/api/bottlenecks":
			h.handleBottlenecks(w, r)
		case "/api/system-info":
			h.handleSystemInfo(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		}
		return
	}

	// Determine the file to serve
	filePath := r.URL.Path
	if filePath == "/" || filePath == "" {
		filePath = "index.html"
	} else {
		// Remove leading slash for fs.Open
		filePath = strings.TrimPrefix(filePath, "/")
	}

	// Try to open the file from embedded FS
	file, err := h.staticFS.Open(filePath)
	if err != nil {
		// Try index.html as fallback for SPA routing
		file, err = h.staticFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		filePath = "index.html"
	}
	defer file.Close()

	// Get file info for content type
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content type based on file extension
	switch {
	case strings.HasSuffix(filePath, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(filePath, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(filePath, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}

	// Serve the file content
	http.ServeContent(w, r, filePath, stat.ModTime(), file.(io.ReadSeeker))
}

// handleAPIRequests serves the JSON API for requests
func (h *Handler) handleAPIRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requests := h.store.GetAll()
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(requests)
}

// handleClearRequests clears all the stored requests
func (h *Handler) handleClearRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Clear the requests in the store
	if err := h.store.Clear(); err != nil {
		http.Error(w, "Error clearing requests", http.StatusInternalServerError)
		return
	}

	// In a real implementation, we would clear the store
	// For now just respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

// handleSSE handles Server-Sent Events for live updates
func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	requests := h.store.GetAll()
	data, _ := json.Marshal(requests)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			requests := h.store.GetAll()
			data, _ := json.Marshal(requests)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// handleCompareRequests serves the JSON API for comparing specific requests
func (h *Handler) handleCompareRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get request IDs from query parameters
	ids := r.URL.Query()["id"]
	if len(ids) < 2 {
		http.Error(w, "At least two request IDs are required", http.StatusBadRequest)
		return
	}

	// Get all requests
	allRequests := h.store.GetAll()

	// Filter requests by IDs
	compareRequests := []interface{}{}
	for _, req := range allRequests {
		for _, id := range ids {
			if req.ID == id {
				compareRequests = append(compareRequests, req)
				break
			}
		}
	}

	// Return the filtered requests
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(compareRequests)
}

// handleReplayRequest handles replaying a captured request
func (h *Handler) handleReplayRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	decoder := json.NewDecoder(r.Body)
	var replayRequest struct {
		RequestID string            `json:"requestId"`
		URL       string            `json:"url"`
		Method    string            `json:"method"`
		Headers   map[string]string `json:"headers"`
		Body      string            `json:"body"`
	}

	if err := decoder.Decode(&replayRequest); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest(replayRequest.Method, replayRequest.URL, strings.NewReader(replayRequest.Body))
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add headers
	for key, value := range replayRequest.Headers {
		req.Header.Add(key, value)
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		http.Error(w, "Error executing request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert headers to map for JSON response
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	// Create response
	replayResponse := struct {
		StatusCode      int                 `json:"statusCode"`
		Headers         map[string][]string `json:"headers"`
		Body            string              `json:"body"`
		Duration        int64               `json:"duration"`
		OriginalRequest string              `json:"originalRequest"`
	}{
		StatusCode:      resp.StatusCode,
		Headers:         headers,
		Body:            string(respBody),
		Duration:        duration,
		OriginalRequest: replayRequest.RequestID,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(replayResponse); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleMetrics serves performance metrics for a specific request
func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestID := r.URL.Query().Get("id")
	if requestID == "" {
		// Return all metrics
		if h.profiler == nil {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"error":"Profiling is not enabled"}`))
			return
		}

		metrics := h.profiler.GetAllMetrics()
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(metrics)
		return
	}

	// Get specific request metrics
	if h.profiler == nil {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":"Profiling is not enabled"}`))
		return
	}

	metrics, found := h.profiler.GetMetrics(requestID)
	if !found {
		// Try to get from request log
		reqLog, found := h.store.Get(requestID)
		if !found || reqLog.PerformanceMetrics == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Metrics not found"}`))
			return
		}
		metrics = reqLog.PerformanceMetrics
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(metrics)
}

// handleFlameGraph generates and serves flame graph data
func (h *Handler) handleFlameGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestID := r.URL.Query().Get("id")
	if requestID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Request ID is required"}`))
		return
	}

	// Get request metrics
	var metrics *profiling.Metrics
	if h.profiler != nil {
		m, found := h.profiler.GetMetrics(requestID)
		if found {
			metrics = m
		}
	}

	if metrics == nil {
		// Try to get from request log
		reqLog, found := h.store.Get(requestID)
		if !found || reqLog.PerformanceMetrics == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Metrics not found for this request"}`))
			return
		}
		metrics = reqLog.PerformanceMetrics
	}

	// Generate flame graph from CPU profile
	if len(metrics.CPUProfile) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"No CPU profile data available"}`))
		return
	}

	flameGraph, err := profiling.GenerateFlameGraph(metrics.CPUProfile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":"Failed to generate flame graph: %v"}`, err)))
		return
	}

	// Convert to D3 format
	d3Data := flameGraph.ConvertToD3Format()

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(d3Data)
}

// handleBottlenecks serves performance bottleneck analysis
func (h *Handler) handleBottlenecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get all requests with metrics
	allRequests := h.store.GetAll()

	type BottleneckSummary struct {
		RequestID   string                 `json:"request_id"`
		Path        string                 `json:"path"`
		Method      string                 `json:"method"`
		Duration    int64                  `json:"duration"`
		Bottlenecks []profiling.Bottleneck `json:"bottlenecks"`
	}

	var summaries []BottleneckSummary

	for _, req := range allRequests {
		if req.PerformanceMetrics != nil && len(req.PerformanceMetrics.Bottlenecks) > 0 {
			summaries = append(summaries, BottleneckSummary{
				RequestID:   req.ID,
				Path:        req.Path,
				Method:      req.Method,
				Duration:    req.Duration,
				Bottlenecks: req.PerformanceMetrics.Bottlenecks,
			})
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(summaries)
}

// handleSystemInfo serves system information for the environment page
func (h *Handler) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get hostname
	hostname, _ := os.Hostname()

	// Get environment variables (filter sensitive ones)
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Redact sensitive environment variables
			if isSensitiveEnvVar(key) {
				value = "[REDACTED]"
			}
			envVars[key] = value
		}
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := map[string]interface{}{
		"goVersion":   runtime.Version(),
		"goos":        runtime.GOOS,
		"goarch":      runtime.GOARCH,
		"hostname":    hostname,
		"cpuCores":    runtime.NumCPU(),
		"memoryUsed":  memStats.Alloc / 1024 / 1024, // Convert to MB
		"memoryTotal": memStats.Sys / 1024 / 1024,   // Convert to MB
		"envVars":     envVars,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(systemInfo)
}

// isSensitiveEnvVar checks if an environment variable key is sensitive
func isSensitiveEnvVar(key string) bool {
	sensitivePatterns := []string{
		"API", "KEY", "SECRET", "PASSWORD", "TOKEN",
		"CREDENTIAL", "AUTH", "PRIVATE", "CERT",
	}

	upperKey := strings.ToUpper(key)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(upperKey, pattern) {
			return true
		}
	}

	return false
}
