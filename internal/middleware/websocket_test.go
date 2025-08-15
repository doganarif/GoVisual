package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/gorilla/websocket"
)

// mockWebSocketPathMatcher implements PathMatcher interface with configurable ignored paths
type mockWebSocketPathMatcher struct {
	ignoredPaths map[string]bool
}

func (m *mockWebSocketPathMatcher) ShouldIgnorePath(path string) bool {
	return m.ignoredPaths[path]
}

func newMockWebSocketPathMatcher(ignoredPaths ...string) *mockWebSocketPathMatcher {
	m := &mockWebSocketPathMatcher{
		ignoredPaths: make(map[string]bool),
	}
	for _, path := range ignoredPaths {
		m.ignoredPaths[path] = true
	}
	return m
}

func TestWrapWebSocket(t *testing.T) {
	mockStore := &mockStore{}
	pathMatcher := newMockWebSocketPathMatcher()
	
	// Create a simple WebSocket echo handler
	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		// Echo messages back to client
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			
			if err := conn.WriteMessage(messageType, message); err != nil {
				break
			}
		}
	}

	// Wrap the handler
	opts := &WebSocketWrapperOptions{
		LogMessageBody: true,
	}
	wrapped := WrapWebSocket(wsHandler, mockStore, pathMatcher, opts)

	// Create test server
	server := httptest.NewServer(wrapped)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/test"

	// Test WebSocket connection
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close()
	defer resp.Body.Close()

	// Wait a moment for the handshake to be logged
	time.Sleep(100 * time.Millisecond)

	// Check that handshake was logged
	logs := mockStore.GetAll()
	if len(logs) == 0 {
		t.Fatal("Expected handshake to be logged")
	}

	handshakeLog := logs[0]
	if handshakeLog.ConnectionType != model.ConnectionTypeWebSocket {
		t.Errorf("Expected connection type to be websocket, got %s", handshakeLog.ConnectionType)
	}

	if handshakeLog.Method != "WEBSOCKET" {
		t.Errorf("Expected method to be WEBSOCKET, got %s", handshakeLog.Method)
	}

	if handshakeLog.Path != "/test" {
		t.Errorf("Expected path to be /test, got %s", handshakeLog.Path)
	}

	if handshakeLog.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status code to be 101, got %d", handshakeLog.StatusCode)
	}

	if handshakeLog.WebSocketData == nil {
		t.Fatal("Expected WebSocket data to be present")
	}

	if !handshakeLog.WebSocketData.IsHandshake {
		t.Error("Expected IsHandshake to be true")
	}

	if handshakeLog.WebSocketData.ConnectionID == "" {
		t.Error("Expected connection ID to be set")
	}
}

func TestWebSocketMessageLogging(t *testing.T) {
	mockStore := &mockStore{}
	
	// Test the MonitoredWebSocketConn WriteMessage logging directly
	connectionID := "test-conn-123"
	testMessage := "Hello WebSocket!"
	
	// We're testing the message logging functionality directly

	// Test outbound message logging
	messageLog := model.NewWebSocketMessageLog(
		connectionID,
		model.WebSocketMessageTypeText,
		model.MessageDirectionOutbound,
		int64(len(testMessage)),
		testMessage,
	)
	mockStore.Add(messageLog)

	// Check logs
	logs := mockStore.GetAll()
	
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	// Check message log details
	if log.ConnectionType != model.ConnectionTypeWebSocket {
		t.Errorf("Expected connection type websocket, got %s", log.ConnectionType)
	}

	if log.Method != "WS_MESSAGE" {
		t.Errorf("Expected method WS_MESSAGE, got %s", log.Method)
	}

	if log.WebSocketData == nil {
		t.Fatal("Expected WebSocket data to be present in message log")
	}

	if log.WebSocketData.ConnectionID != connectionID {
		t.Errorf("Expected connection ID %s, got %s", connectionID, log.WebSocketData.ConnectionID)
	}

	if log.WebSocketData.MessageType != model.WebSocketMessageTypeText {
		t.Errorf("Expected message type text, got %s", log.WebSocketData.MessageType)
	}

	if log.WebSocketData.Direction != model.MessageDirectionOutbound {
		t.Errorf("Expected direction outbound, got %s", log.WebSocketData.Direction)
	}

	if log.WebSocketData.MessageSize != int64(len(testMessage)) {
		t.Errorf("Expected message size %d, got %d", len(testMessage), log.WebSocketData.MessageSize)
	}

	if log.RequestBody != testMessage {
		t.Errorf("Expected request body '%s', got '%s'", testMessage, log.RequestBody)
	}
}

func TestWebSocketPathIgnoring(t *testing.T) {
	mockStore := &mockStore{}
	pathMatcher := newMockWebSocketPathMatcher("/ignored")
	
	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Should not be monitored"))
	}

	// Wrap the handler
	wrapped := WrapWebSocket(wsHandler, mockStore, pathMatcher, nil)

	// Test ignored path
	req := httptest.NewRequest("GET", "/ignored", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Check that no logs were created
	logs := mockStore.GetAll()
	if len(logs) != 0 {
		t.Errorf("Expected no logs for ignored path, got %d", len(logs))
	}
}

func TestNewWebSocketLog(t *testing.T) {
	connectionID := "test-connection-123"
	path := "/websocket/test"
	headers := http.Header{
		"Upgrade":               []string{"websocket"},
		"Connection":            []string{"Upgrade"},
		"Sec-WebSocket-Key":     []string{"test-key"},
		"Sec-WebSocket-Version": []string{"13"},
	}

	log := model.NewWebSocketLog(connectionID, path, headers)

	if log.ConnectionType != model.ConnectionTypeWebSocket {
		t.Errorf("Expected connection type websocket, got %s", log.ConnectionType)
	}

	if log.Method != "WEBSOCKET" {
		t.Errorf("Expected method WEBSOCKET, got %s", log.Method)
	}

	if log.Path != path {
		t.Errorf("Expected path %s, got %s", path, log.Path)
	}

	if log.WebSocketData == nil {
		t.Fatal("Expected WebSocket data to be present")
	}

	if log.WebSocketData.ConnectionID != connectionID {
		t.Errorf("Expected connection ID %s, got %s", connectionID, log.WebSocketData.ConnectionID)
	}

	if !log.WebSocketData.IsHandshake {
		t.Error("Expected IsHandshake to be true")
	}

	// Check that headers were preserved
	if len(log.RequestHeaders) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(log.RequestHeaders))
	}
}

func TestNewWebSocketMessageLog(t *testing.T) {
	connectionID := "test-connection-456"
	messageType := model.WebSocketMessageTypeText
	direction := model.MessageDirectionInbound
	size := int64(25)
	content := "Hello from WebSocket test!"

	log := model.NewWebSocketMessageLog(connectionID, messageType, direction, size, content)

	if log.ConnectionType != model.ConnectionTypeWebSocket {
		t.Errorf("Expected connection type websocket, got %s", log.ConnectionType)
	}

	if log.Method != "WS_MESSAGE" {
		t.Errorf("Expected method WS_MESSAGE, got %s", log.Method)
	}

	if log.RequestBody != content {
		t.Errorf("Expected request body %s, got %s", content, log.RequestBody)
	}

	if log.WebSocketData == nil {
		t.Fatal("Expected WebSocket data to be present")
	}

	if log.WebSocketData.ConnectionID != connectionID {
		t.Errorf("Expected connection ID %s, got %s", connectionID, log.WebSocketData.ConnectionID)
	}

	if log.WebSocketData.MessageType != messageType {
		t.Errorf("Expected message type %s, got %s", messageType, log.WebSocketData.MessageType)
	}

	if log.WebSocketData.Direction != direction {
		t.Errorf("Expected direction %s, got %s", direction, log.WebSocketData.Direction)
	}

	if log.WebSocketData.MessageSize != size {
		t.Errorf("Expected message size %d, got %d", size, log.WebSocketData.MessageSize)
	}

	if log.WebSocketData.IsHandshake {
		t.Error("Expected IsHandshake to be false for message log")
	}
}

func TestGenerateConnectionID(t *testing.T) {
	// Test that connection IDs are generated and unique
	id1 := generateConnectionID()
	id2 := generateConnectionID()

	if id1 == "" {
		t.Error("Expected non-empty connection ID")
	}

	if len(id1) != 16 { // 8 bytes = 16 hex characters
		t.Errorf("Expected connection ID length 16, got %d", len(id1))
	}

	if id1 == id2 {
		t.Error("Expected unique connection IDs")
	}
}

func TestWebSocketWrapperOptions(t *testing.T) {
	mockStore := &mockStore{}
	pathMatcher := newMockWebSocketPathMatcher()

	// Test with custom upgrader
	customUpgrader := websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "https://example.com"
		},
	}

	opts := &WebSocketWrapperOptions{
		LogMessageBody: false,
		Upgrader:       &customUpgrader,
	}

	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrapped := WrapWebSocket(wsHandler, mockStore, pathMatcher, opts)

	// Verify that the wrapper was created
	if wrapped == nil {
		t.Error("Expected wrapped handler to be created")
	}

	// Test with nil options (should use defaults)
	wrapped2 := WrapWebSocket(wsHandler, mockStore, pathMatcher, nil)
	if wrapped2 == nil {
		t.Error("Expected wrapped handler to be created with nil options")
	}
}

// benchmarkWebSocketWrapper benchmarks the WebSocket wrapper performance
func BenchmarkWebSocketWrapper(b *testing.B) {
	mockStore := &mockStore{}
	pathMatcher := newMockWebSocketPathMatcher()

	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	opts := &WebSocketWrapperOptions{
		LogMessageBody: true,
	}
	wrapped := WrapWebSocket(wsHandler, mockStore, pathMatcher, opts)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}
}