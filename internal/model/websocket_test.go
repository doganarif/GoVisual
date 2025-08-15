package model

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestConnectionTypeConstants(t *testing.T) {
	// Test connection type constants
	if ConnectionTypeHTTP != "http" {
		t.Errorf("Expected ConnectionTypeHTTP to be 'http', got %s", ConnectionTypeHTTP)
	}

	if ConnectionTypeWebSocket != "websocket" {
		t.Errorf("Expected ConnectionTypeWebSocket to be 'websocket', got %s", ConnectionTypeWebSocket)
	}
}

func TestMessageDirectionConstants(t *testing.T) {
	// Test message direction constants
	if MessageDirectionInbound != "inbound" {
		t.Errorf("Expected MessageDirectionInbound to be 'inbound', got %s", MessageDirectionInbound)
	}

	if MessageDirectionOutbound != "outbound" {
		t.Errorf("Expected MessageDirectionOutbound to be 'outbound', got %s", MessageDirectionOutbound)
	}
}

func TestWebSocketMessageTypeConstants(t *testing.T) {
	// Test WebSocket message type constants
	expectedTypes := map[WebSocketMessageType]string{
		WebSocketMessageTypeText:   "text",
		WebSocketMessageTypeBinary: "binary",
		WebSocketMessageTypePing:   "ping",
		WebSocketMessageTypePong:   "pong",
		WebSocketMessageTypeClose:  "close",
	}

	for constant, expected := range expectedTypes {
		if string(constant) != expected {
			t.Errorf("Expected %s constant to be '%s', got %s", expected, expected, constant)
		}
	}
}

func TestNewRequestLogHTTP(t *testing.T) {
	// Create a mock HTTP request
	req, _ := http.NewRequest("GET", "/api/users?limit=10", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Accept", "application/json")

	// Create RequestLog from HTTP request
	log := NewRequestLog(req)

	// Test basic fields
	if log.ConnectionType != ConnectionTypeHTTP {
		t.Errorf("Expected connection type 'http', got %s", log.ConnectionType)
	}

	if log.Method != "GET" {
		t.Errorf("Expected method 'GET', got %s", log.Method)
	}

	if log.Path != "/api/users" {
		t.Errorf("Expected path '/api/users', got %s", log.Path)
	}

	if log.Query != "limit=10" {
		t.Errorf("Expected query 'limit=10', got %s", log.Query)
	}

	// Test headers
	if log.RequestHeaders.Get("User-Agent") != "test-agent" {
		t.Errorf("Expected User-Agent header to be preserved")
	}

	if log.RequestHeaders.Get("Accept") != "application/json" {
		t.Errorf("Expected Accept header to be preserved")
	}

	// Test WebSocket data is nil for HTTP requests
	if log.WebSocketData != nil {
		t.Error("Expected WebSocketData to be nil for HTTP requests")
	}

	// Test ID generation
	if log.ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Test timestamp
	if log.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewWebSocketLog(t *testing.T) {
	connectionID := "ws-conn-123"
	path := "/chat"
	headers := http.Header{
		"Upgrade":               []string{"websocket"},
		"Connection":            []string{"Upgrade"},
		"Sec-WebSocket-Key":     []string{"dGhlIHNhbXBsZSBub25jZQ=="},
		"Sec-WebSocket-Version": []string{"13"},
		"Origin":                []string{"https://example.com"},
	}

	log := NewWebSocketLog(connectionID, path, headers)

	// Test basic fields
	if log.ConnectionType != ConnectionTypeWebSocket {
		t.Errorf("Expected connection type 'websocket', got %s", log.ConnectionType)
	}

	if log.Method != "WEBSOCKET" {
		t.Errorf("Expected method 'WEBSOCKET', got %s", log.Method)
	}

	if log.Path != path {
		t.Errorf("Expected path '%s', got %s", path, log.Path)
	}

	// Test headers are preserved
	if log.RequestHeaders.Get("Upgrade") != "websocket" {
		t.Errorf("Expected Upgrade header to be preserved, got: %s", log.RequestHeaders.Get("Upgrade"))
	}

	// Access header directly from map since Get() might have canonicalization issues
	secWebSocketKey := log.RequestHeaders["Sec-WebSocket-Key"]
	if len(secWebSocketKey) == 0 || secWebSocketKey[0] != "dGhlIHNhbXBsZSBub25jZQ==" {
		t.Errorf("Expected Sec-WebSocket-Key header to be preserved, got: %v", secWebSocketKey)
	}

	// Test WebSocket data
	if log.WebSocketData == nil {
		t.Fatal("Expected WebSocketData to be present")
	}

	if log.WebSocketData.ConnectionID != connectionID {
		t.Errorf("Expected connection ID '%s', got %s", connectionID, log.WebSocketData.ConnectionID)
	}

	if !log.WebSocketData.IsHandshake {
		t.Error("Expected IsHandshake to be true")
	}

	// Test ID and timestamp generation
	if log.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if log.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewWebSocketMessageLog(t *testing.T) {
	connectionID := "ws-conn-456"
	messageType := WebSocketMessageTypeText
	direction := MessageDirectionInbound
	size := int64(42)
	content := "Hello WebSocket world!"

	log := NewWebSocketMessageLog(connectionID, messageType, direction, size, content)

	// Test basic fields
	if log.ConnectionType != ConnectionTypeWebSocket {
		t.Errorf("Expected connection type 'websocket', got %s", log.ConnectionType)
	}

	if log.Method != "WS_MESSAGE" {
		t.Errorf("Expected method 'WS_MESSAGE', got %s", log.Method)
	}

	if log.RequestBody != content {
		t.Errorf("Expected request body '%s', got %s", content, log.RequestBody)
	}

	// Test WebSocket data
	if log.WebSocketData == nil {
		t.Fatal("Expected WebSocketData to be present")
	}

	if log.WebSocketData.ConnectionID != connectionID {
		t.Errorf("Expected connection ID '%s', got %s", connectionID, log.WebSocketData.ConnectionID)
	}

	if log.WebSocketData.MessageType != messageType {
		t.Errorf("Expected message type '%s', got %s", messageType, log.WebSocketData.MessageType)
	}

	if log.WebSocketData.Direction != direction {
		t.Errorf("Expected direction '%s', got %s", direction, log.WebSocketData.Direction)
	}

	if log.WebSocketData.MessageSize != size {
		t.Errorf("Expected message size %d, got %d", size, log.WebSocketData.MessageSize)
	}

	if log.WebSocketData.IsHandshake {
		t.Error("Expected IsHandshake to be false for message logs")
	}

	// Test ID and timestamp generation
	if log.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if log.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestWebSocketDataJSONSerialization(t *testing.T) {
	// Create a complete WebSocket data structure
	wsData := &WebSocketData{
		ConnectionID:  "conn-789",
		MessageType:   WebSocketMessageTypeBinary,
		Direction:     MessageDirectionOutbound,
		MessageSize:   1024,
		CloseCode:     1000,
		CloseReason:   "Normal closure",
		Subprotocol:   "chat",
		Extensions:    []string{"permessage-deflate", "x-webkit-deflate-frame"},
		IsHandshake:   false,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(wsData)
	if err != nil {
		t.Fatalf("Failed to marshal WebSocketData: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled WebSocketData
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal WebSocketData: %v", err)
	}

	// Verify all fields
	if unmarshaled.ConnectionID != wsData.ConnectionID {
		t.Errorf("Connection ID mismatch: expected %s, got %s", wsData.ConnectionID, unmarshaled.ConnectionID)
	}

	if unmarshaled.MessageType != wsData.MessageType {
		t.Errorf("Message type mismatch: expected %s, got %s", wsData.MessageType, unmarshaled.MessageType)
	}

	if unmarshaled.Direction != wsData.Direction {
		t.Errorf("Direction mismatch: expected %s, got %s", wsData.Direction, unmarshaled.Direction)
	}

	if unmarshaled.MessageSize != wsData.MessageSize {
		t.Errorf("Message size mismatch: expected %d, got %d", wsData.MessageSize, unmarshaled.MessageSize)
	}

	if unmarshaled.CloseCode != wsData.CloseCode {
		t.Errorf("Close code mismatch: expected %d, got %d", wsData.CloseCode, unmarshaled.CloseCode)
	}

	if unmarshaled.CloseReason != wsData.CloseReason {
		t.Errorf("Close reason mismatch: expected %s, got %s", wsData.CloseReason, unmarshaled.CloseReason)
	}

	if unmarshaled.Subprotocol != wsData.Subprotocol {
		t.Errorf("Subprotocol mismatch: expected %s, got %s", wsData.Subprotocol, unmarshaled.Subprotocol)
	}

	if len(unmarshaled.Extensions) != len(wsData.Extensions) {
		t.Errorf("Extensions length mismatch: expected %d, got %d", len(wsData.Extensions), len(unmarshaled.Extensions))
	}

	if unmarshaled.IsHandshake != wsData.IsHandshake {
		t.Errorf("IsHandshake mismatch: expected %t, got %t", wsData.IsHandshake, unmarshaled.IsHandshake)
	}
}

func TestRequestLogJSONSerialization(t *testing.T) {
	// Create a complete RequestLog with WebSocket data
	log := &RequestLog{
		ID:             "test-log-123",
		Timestamp:      time.Now().UTC(),
		ConnectionType: ConnectionTypeWebSocket,
		Method:         "WS_MESSAGE",
		Path:           "/websocket",
		Query:          "room=test",
		RequestHeaders: http.Header{
			"Upgrade":    []string{"websocket"},
			"Connection": []string{"Upgrade"},
		},
		ResponseHeaders: http.Header{
			"Upgrade":    []string{"websocket"},
			"Connection": []string{"Upgrade"},
		},
		StatusCode:   101,
		Duration:     250,
		RequestBody:  "Hello WebSocket!",
		ResponseBody: "",
		Error:        "",
		WebSocketData: &WebSocketData{
			ConnectionID: "ws-conn-test",
			MessageType:  WebSocketMessageTypeText,
			Direction:    MessageDirectionInbound,
			MessageSize:  16,
			IsHandshake:  false,
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal RequestLog: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled RequestLog
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal RequestLog: %v", err)
	}

	// Verify key fields
	if unmarshaled.ID != log.ID {
		t.Errorf("ID mismatch: expected %s, got %s", log.ID, unmarshaled.ID)
	}

	if unmarshaled.ConnectionType != log.ConnectionType {
		t.Errorf("Connection type mismatch: expected %s, got %s", log.ConnectionType, unmarshaled.ConnectionType)
	}

	if unmarshaled.Method != log.Method {
		t.Errorf("Method mismatch: expected %s, got %s", log.Method, unmarshaled.Method)
	}

	// Verify WebSocket data
	if unmarshaled.WebSocketData == nil {
		t.Fatal("Expected WebSocket data to be present after unmarshaling")
	}

	if unmarshaled.WebSocketData.ConnectionID != log.WebSocketData.ConnectionID {
		t.Errorf("WebSocket connection ID mismatch: expected %s, got %s", 
			log.WebSocketData.ConnectionID, unmarshaled.WebSocketData.ConnectionID)
	}

	if unmarshaled.WebSocketData.MessageType != log.WebSocketData.MessageType {
		t.Errorf("WebSocket message type mismatch: expected %s, got %s", 
			log.WebSocketData.MessageType, unmarshaled.WebSocketData.MessageType)
	}
}

func TestWebSocketDataOmitEmpty(t *testing.T) {
	// Create a RequestLog without WebSocket data (HTTP request)
	log := &RequestLog{
		ID:             "http-log-123",
		Timestamp:      time.Now().UTC(),
		ConnectionType: ConnectionTypeHTTP,
		Method:         "GET",
		Path:           "/api/test",
		StatusCode:     200,
		Duration:       100,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Failed to marshal RequestLog: %v", err)
	}

	// Verify that WebSocketData field is omitted when nil
	jsonStr := string(jsonData)
	if contains(jsonStr, "WebSocketData") {
		t.Error("Expected WebSocketData field to be omitted when nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs and ensure they are unique and properly formatted
	ids := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		id := generateID()
		
		// Check that ID is not empty
		if id == "" {
			t.Error("Generated ID should not be empty")
		}
		
		// Check that ID is unique
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true
		
		// Check ID format (should be timestamp-based)
		if len(id) < 15 { // Minimum expected length for timestamp format
			t.Errorf("Generated ID seems too short: %s", id)
		}
	}
}

func BenchmarkNewRequestLog(b *testing.B) {
	req, _ := http.NewRequest("GET", "/api/benchmark", nil)
	req.Header.Set("User-Agent", "benchmark-test")
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = NewRequestLog(req)
	}
}

func BenchmarkNewWebSocketLog(b *testing.B) {
	headers := http.Header{
		"Upgrade":    []string{"websocket"},
		"Connection": []string{"Upgrade"},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = NewWebSocketLog("conn-123", "/ws", headers)
	}
}

func BenchmarkNewWebSocketMessageLog(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = NewWebSocketMessageLog("conn-123", WebSocketMessageTypeText, MessageDirectionInbound, 25, "test message")
	}
}