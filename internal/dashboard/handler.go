package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/store"
)

// Handler is the HTTP handler for the dashboard
type Handler struct {
	store    store.Store
	template *template.Template
}

// NewHandler creates a new dashboard handler
func NewHandler(store store.Store) *Handler {
	// Parse templates from embedded filesystem
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
	template.Must(tmpl.ParseFS(templateFS, "templates/components/*.html"))

	return &Handler{
		store:    store,
		template: tmpl,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// API endpoints
	switch path.Clean(r.URL.Path) {
	case "/api/requests":
		h.handleAPIRequests(w, r)
		return
	case "/api/events":
		h.handleSSE(w, r)
		return
	case "/api/clear":
		h.handleClearRequests(w, r)
		return
	case "/api/compare":
		h.handleCompareRequests(w, r)
		return
	case "/api/replay":
		h.handleReplayRequest(w, r)
		return
	case "/":
		h.handleDashboard(w, r)
		return
	default:
		// Serve a simple 404 page
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	}
}

// handleDashboard serves the dashboard HTML
func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Get system and environment information
	hostname, _ := os.Hostname()

	// Filter environment variables to avoid exposing sensitive information
	filteredEnvVars := make(map[string]string)
	for _, env := range os.Environ() {
		// Split env var into key and value
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				key := env[:i]
				value := env[i+1:]

				// Skip sensitive environment variables
				if isSensitiveEnvVar(key) {
					filteredEnvVars[key] = "[REDACTED]"
				} else {
					filteredEnvVars[key] = value
				}
				break
			}
		}
	}

	// Get the data to pass to the template
	data := map[string]interface{}{
		"Requests":  h.store.GetAll(),
		"GoVersion": runtime.Version(),
		"GOOS":      runtime.GOOS,
		"GOARCH":    runtime.GOARCH,
		"Hostname":  hostname,
		"OS":        runtime.GOOS,
		"CPUs":      runtime.NumCPU(),
		"EnvVars":   filteredEnvVars,
	}

	// Execute the dashboard template
	err := h.template.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
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

	// In a real implementation, we would clear the store
	// For now just respond with success
	w.Header().Set("Content-Type", "application/json")
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

// isSensitiveEnvVar returns true if the environment variable key is considered sensitive
func isSensitiveEnvVar(key string) bool {
	sensitiveKeys := []string{
		"API_KEY", "SECRET", "PASSWORD", "TOKEN", "CREDENTIAL", "AUTH",
		"CERTIFICATE", "PRIVATE", "KEY",
	}

	key = upper(key)
	for _, sensitive := range sensitiveKeys {
		if contains(key, sensitive) {
			return true
		}
	}

	return false
}

// upper converts a string to uppercase
func upper(s string) string {
	result := []byte(s)
	for i := 0; i < len(result); i++ {
		if 'a' <= result[i] && result[i] <= 'z' {
			result[i] -= 'a' - 'A'
		}
	}
	return string(result)
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
