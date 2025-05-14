package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/store"
)

// AgentAPI handles requests from remote agents.
type AgentAPI struct {
	store store.Store
}

// NewAgentAPI creates a new API handler for agent requests.
func NewAgentAPI(store store.Store) *AgentAPI {
	return &AgentAPI{
		store: store,
	}
}

// RegisterHandlers registers HTTP handlers for agent API endpoints.
func (api *AgentAPI) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/agent/logs", api.handleLogs)
	mux.HandleFunc("/api/agent/logs/batch", api.handleBatchLogs)
	mux.HandleFunc("/api/agent/status", api.handleStatus)
}

// handleLogs handles requests to add a single request log.
func (api *AgentAPI) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request log
	var reqLog model.RequestLog
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqLog); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Store the log
	api.store.Add(&reqLog)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleBatchLogs handles requests to add multiple request logs in a batch.
func (api *AgentAPI) handleBatchLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode batch of request logs
	var reqLogs []*model.RequestLog
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqLogs); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Store each log
	for _, log := range reqLogs {
		api.store.Add(log)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"count":  len(reqLogs),
	})
}

// handleStatus handles agent status check requests.
func (api *AgentAPI) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get agent ID from query params
	agentID := r.URL.Query().Get("id")
	if agentID == "" {
		http.Error(w, "Missing agent ID", http.StatusBadRequest)
		return
	}

	// Respond with server status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"agent_id":  agentID,
		"timestamp": serverTimestamp(),
		"version":   "1.0.0", // TODO: Get from app version
	})
}

// serverTimestamp returns the current server timestamp in ISO 8601 format.
func serverTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
