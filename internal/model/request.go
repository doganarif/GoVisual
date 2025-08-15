package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// ConnectionType defines the type of connection being logged
type ConnectionType string

const (
	ConnectionTypeHTTP      ConnectionType = "http"
	ConnectionTypeWebSocket ConnectionType = "websocket"
)

// MessageDirection defines the direction of WebSocket messages
type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

// WebSocketMessageType defines the type of WebSocket message
type WebSocketMessageType string

const (
	WebSocketMessageTypeText   WebSocketMessageType = "text"
	WebSocketMessageTypeBinary WebSocketMessageType = "binary"
	WebSocketMessageTypePing   WebSocketMessageType = "ping"
	WebSocketMessageTypePong   WebSocketMessageType = "pong"
	WebSocketMessageTypeClose  WebSocketMessageType = "close"
)

// WebSocketData contains WebSocket-specific information
type WebSocketData struct {
	ConnectionID  string                `json:"ConnectionID,omitempty"`
	MessageType   WebSocketMessageType  `json:"MessageType,omitempty"`
	Direction     MessageDirection      `json:"Direction,omitempty"`
	MessageSize   int64                 `json:"MessageSize,omitempty"`
	CloseCode     int                   `json:"CloseCode,omitempty"`
	CloseReason   string                `json:"CloseReason,omitempty"`
	Subprotocol   string                `json:"Subprotocol,omitempty"`
	Extensions    []string              `json:"Extensions,omitempty"`
	IsHandshake   bool                  `json:"IsHandshake,omitempty"`
}

type RequestLog struct {
	ID              string                   `json:"ID" bson:"_id"`
	Timestamp       time.Time                `json:"Timestamp"`
	ConnectionType  ConnectionType           `json:"ConnectionType"`
	Method          string                   `json:"Method"`
	Path            string                   `json:"Path"`
	Query           string                   `json:"Query"`
	RequestHeaders  http.Header              `json:"RequestHeaders"`
	ResponseHeaders http.Header              `json:"ResponseHeaders"`
	StatusCode      int                      `json:"StatusCode"`
	Duration        int64                    `json:"Duration"`
	RequestBody     string                   `json:"RequestBody,omitempty"`
	ResponseBody    string                   `json:"ResponseBody,omitempty"`
	Error           string                   `json:"Error,omitempty"`
	MiddlewareTrace []map[string]interface{} `json:"MiddlewareTrace,omitempty"`
	RouteTrace      map[string]interface{}   `json:"RouteTrace,omitempty"`
	
	// WebSocket-specific fields
	WebSocketData   *WebSocketData           `json:"WebSocketData,omitempty"`
}

func NewRequestLog(req *http.Request) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		ConnectionType: ConnectionTypeHTTP,
		Method:         req.Method,
		Path:           req.URL.Path,
		Query:          req.URL.RawQuery,
		RequestHeaders: req.Header,
	}
}

// NewWebSocketLog creates a new RequestLog for WebSocket connections
func NewWebSocketLog(connectionID, path string, headers http.Header) *RequestLog {
	// Create a copy of the headers to avoid issues with shared references
	headersCopy := make(http.Header)
	for key, values := range headers {
		headersCopy[key] = append([]string(nil), values...)
	}

	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		ConnectionType: ConnectionTypeWebSocket,
		Method:         "WEBSOCKET",
		Path:           path,
		RequestHeaders: headersCopy,
		WebSocketData: &WebSocketData{
			ConnectionID: connectionID,
			IsHandshake:  true,
		},
	}
}

// NewWebSocketMessageLog creates a new RequestLog for WebSocket messages
func NewWebSocketMessageLog(connectionID string, messageType WebSocketMessageType, direction MessageDirection, size int64, content string) *RequestLog {
	return &RequestLog{
		ID:             generateID(),
		Timestamp:      time.Now(),
		ConnectionType: ConnectionTypeWebSocket,
		Method:         "WS_MESSAGE",
		RequestBody:    content,
		WebSocketData: &WebSocketData{
			ConnectionID: connectionID,
			MessageType:  messageType,
			Direction:    direction,
			MessageSize:  size,
		},
	}
}

func generateID() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(randomBytes))
}
