package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/store"
	"github.com/nats-io/nats.go"
)

// NATSHandler handles agent messages received via NATS.
type NATSHandler struct {
	store store.Store
	conn  *nats.Conn
	subs  []*nats.Subscription
}

// NewNATSHandler creates a new NATS handler for agent messages.
func NewNATSHandler(store store.Store, serverURL string, opts ...nats.Option) (*NATSHandler, error) {
	// Connect to NATS
	conn, err := nats.Connect(serverURL, opts...)
	if err != nil {
		return nil, err
	}

	return &NATSHandler{
		store: store,
		conn:  conn,
		subs:  make([]*nats.Subscription, 0),
	}, nil
}

// Start begins listening for agent messages on NATS.
func (h *NATSHandler) Start() error {
	// Subscribe to single log messages
	singleSub, err := h.conn.Subscribe("govisual.logs.single", h.handleSingleLog)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, singleSub)

	// Subscribe to batch log messages
	batchSub, err := h.conn.Subscribe("govisual.logs.batch", h.handleBatchLogs)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, batchSub)

	// Subscribe to agent status messages
	statusSub, err := h.conn.Subscribe("govisual.agent.status", h.handleAgentStatus)
	if err != nil {
		return err
	}
	h.subs = append(h.subs, statusSub)

	log.Println("NATS handler started and listening for agent messages")
	return nil
}

// Stop unsubscribes from all NATS channels and closes the connection.
func (h *NATSHandler) Stop() error {
	// Unsubscribe from all subscriptions
	for _, sub := range h.subs {
		if sub != nil {
			if err := sub.Unsubscribe(); err != nil {
				log.Printf("Error unsubscribing from NATS: %v", err)
			}
		}
	}

	// Close the connection
	h.conn.Close()
	log.Println("NATS handler stopped")
	return nil
}

// handleSingleLog processes a single log message from an agent.
func (h *NATSHandler) handleSingleLog(msg *nats.Msg) {
	var reqLog model.RequestLog
	if err := json.Unmarshal(msg.Data, &reqLog); err != nil {
		log.Printf("Error unmarshaling log message: %v", err)
		return
	}

	// Store the log
	h.store.Add(&reqLog)

	// Acknowledge message receipt
	if msg.Reply != "" {
		response := map[string]string{
			"status":    "success",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}
		h.conn.Publish(msg.Reply, data)
	}
}

// handleBatchLogs processes a batch of log messages from an agent.
func (h *NATSHandler) handleBatchLogs(msg *nats.Msg) {
	var reqLogs []*model.RequestLog
	if err := json.Unmarshal(msg.Data, &reqLogs); err != nil {
		log.Printf("Error unmarshaling batch log message: %v", err)
		return
	}

	// Store each log
	for _, log := range reqLogs {
		h.store.Add(log)
	}

	// Acknowledge message receipt
	if msg.Reply != "" {
		response := map[string]interface{}{
			"status":    "success",
			"count":     len(reqLogs),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}
		h.conn.Publish(msg.Reply, data)
	}
}

// handleAgentStatus processes agent status messages.
func (h *NATSHandler) handleAgentStatus(msg *nats.Msg) {
	var status struct {
		AgentID   string `json:"agent_id"`
		AgentType string `json:"agent_type"`
		Hostname  string `json:"hostname"`
		Version   string `json:"version"`
	}

	if err := json.Unmarshal(msg.Data, &status); err != nil {
		log.Printf("Error unmarshaling agent status message: %v", err)
		return
	}

	// Log agent status (could be stored in a registry for monitoring)
	log.Printf("Agent status received: %s (%s) on %s, version %s",
		status.AgentID, status.AgentType, status.Hostname, status.Version)

	// Reply with server status
	if msg.Reply != "" {
		response := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0", // TODO: Get from app version
		}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			return
		}
		h.conn.Publish(msg.Reply, data)
	}
}
