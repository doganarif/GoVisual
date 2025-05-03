package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
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

	// Get the data to pass to the template
	data := map[string]interface{}{
		"Requests": h.store.GetAll(),
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
