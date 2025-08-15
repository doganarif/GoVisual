package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/store"
	"github.com/gorilla/websocket"
)

// WebSocketUpgrader is the default WebSocket upgrader with origin checking disabled for development
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

// WebSocketWrapper wraps a WebSocket handler with monitoring functionality
type WebSocketWrapper struct {
	store           store.Store
	pathMatcher     PathMatcher
	logMessageBody  bool
	upgrader        websocket.Upgrader
	handler         http.HandlerFunc
}

// WebSocketWrapperOptions holds configuration options for WebSocket wrapper
type WebSocketWrapperOptions struct {
	LogMessageBody bool
	Upgrader       *websocket.Upgrader
}

// WrapWebSocket creates a wrapper for WebSocket connections with monitoring capabilities
func WrapWebSocket(handler http.HandlerFunc, store store.Store, pathMatcher PathMatcher, opts *WebSocketWrapperOptions) http.Handler {
	wrapper := &WebSocketWrapper{
		store:       store,
		pathMatcher: pathMatcher,
		handler:     handler,
	}

	// Apply options
	if opts != nil {
		wrapper.logMessageBody = opts.LogMessageBody
		if opts.Upgrader != nil {
			wrapper.upgrader = *opts.Upgrader
		} else {
			wrapper.upgrader = WebSocketUpgrader
		}
	} else {
		wrapper.upgrader = WebSocketUpgrader
	}

	return wrapper
}

// ServeHTTP implements the http.Handler interface
func (w *WebSocketWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Check if the path should be ignored
	if w.pathMatcher != nil && w.pathMatcher.ShouldIgnorePath(r.URL.Path) {
		w.handler.ServeHTTP(rw, r)
		return
	}

	// Generate connection ID
	connectionID := generateConnectionID()

	// Create handshake log
	handshakeLog := model.NewWebSocketLog(connectionID, r.URL.Path, r.Header)
	handshakeLog.Query = r.URL.RawQuery

	start := time.Now()

	// Upgrade the HTTP connection to WebSocket
	conn, err := w.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		// Log failed handshake
		handshakeLog.Error = err.Error()
		handshakeLog.StatusCode = http.StatusBadRequest
		handshakeLog.Duration = time.Since(start).Milliseconds()
		w.store.Add(handshakeLog)
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Log successful handshake
	handshakeLog.StatusCode = http.StatusSwitchingProtocols
	handshakeLog.Duration = time.Since(start).Milliseconds()
	if conn.Subprotocol() != "" {
		handshakeLog.WebSocketData.Subprotocol = conn.Subprotocol()
	}
	w.store.Add(handshakeLog)

	// Wrap the connection for monitoring
	monitoredConn := &MonitoredWebSocketConn{
		Conn:           conn,
		connectionID:   connectionID,
		store:          w.store,
		logMessageBody: w.logMessageBody,
	}

	// Create a custom handler that uses the monitored connection
	monitoredHandler := func(rw http.ResponseWriter, r *http.Request) {
		// Call the original handler, but it will use our monitored connection
		// We need to pass the monitored connection somehow
		// For now, we'll handle the connection monitoring directly here
		w.handleWebSocketConnection(monitoredConn, r)
	}

	monitoredHandler(rw, r)
}

// handleWebSocketConnection manages the WebSocket connection lifecycle
func (w *WebSocketWrapper) handleWebSocketConnection(conn *MonitoredWebSocketConn, r *http.Request) {
	// Set up ping/pong handlers
	conn.SetPingHandler(func(message string) error {
		// Log ping message
		pingLog := model.NewWebSocketMessageLog(
			conn.connectionID,
			model.WebSocketMessageTypePing,
			model.MessageDirectionInbound,
			int64(len(message)),
			message,
		)
		w.store.Add(pingLog)

		// Send pong response
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(10*time.Second))
		if err == nil {
			// Log pong response
			pongLog := model.NewWebSocketMessageLog(
				conn.connectionID,
				model.WebSocketMessageTypePong,
				model.MessageDirectionOutbound,
				int64(len(message)),
				message,
			)
			w.store.Add(pongLog)
		}
		return err
	})

	conn.SetPongHandler(func(message string) error {
		// Log pong message
		pongLog := model.NewWebSocketMessageLog(
			conn.connectionID,
			model.WebSocketMessageTypePong,
			model.MessageDirectionInbound,
			int64(len(message)),
			message,
		)
		w.store.Add(pongLog)
		return nil
	})

	// Start monitoring the connection
	go w.monitorConnection(conn)

	// Let the original handler take control
	// Note: In a real implementation, you might want to provide a way
	// for the handler to access the monitored connection
}

// monitorConnection monitors WebSocket messages
func (w *WebSocketWrapper) monitorConnection(conn *MonitoredWebSocketConn) {
	defer func() {
		// Log connection closure
		closeLog := model.NewWebSocketMessageLog(
			conn.connectionID,
			model.WebSocketMessageTypeClose,
			model.MessageDirectionInbound,
			0,
			"Connection closed",
		)
		w.store.Add(closeLog)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Determine WebSocket message type
		var wsMessageType model.WebSocketMessageType
		switch messageType {
		case websocket.TextMessage:
			wsMessageType = model.WebSocketMessageTypeText
		case websocket.BinaryMessage:
			wsMessageType = model.WebSocketMessageTypeBinary
		case websocket.CloseMessage:
			wsMessageType = model.WebSocketMessageTypeClose
		case websocket.PingMessage:
			wsMessageType = model.WebSocketMessageTypePing
		case websocket.PongMessage:
			wsMessageType = model.WebSocketMessageTypePong
		}

		// Prepare message content
		messageContent := ""
		if w.logMessageBody {
			if messageType == websocket.TextMessage {
				messageContent = string(message)
			} else {
				messageContent = hex.EncodeToString(message)
			}
		}

		// Log the message
		messageLog := model.NewWebSocketMessageLog(
			conn.connectionID,
			wsMessageType,
			model.MessageDirectionInbound,
			int64(len(message)),
			messageContent,
		)
		w.store.Add(messageLog)
	}
}

// MonitoredWebSocketConn wraps a WebSocket connection to monitor messages
type MonitoredWebSocketConn struct {
	*websocket.Conn
	connectionID   string
	store          store.Store
	logMessageBody bool
}

// WriteMessage overrides the WriteMessage method to log outbound messages
func (m *MonitoredWebSocketConn) WriteMessage(messageType int, data []byte) error {
	// Determine WebSocket message type
	var wsMessageType model.WebSocketMessageType
	switch messageType {
	case websocket.TextMessage:
		wsMessageType = model.WebSocketMessageTypeText
	case websocket.BinaryMessage:
		wsMessageType = model.WebSocketMessageTypeBinary
	case websocket.CloseMessage:
		wsMessageType = model.WebSocketMessageTypeClose
	case websocket.PingMessage:
		wsMessageType = model.WebSocketMessageTypePing
	case websocket.PongMessage:
		wsMessageType = model.WebSocketMessageTypePong
	}

	// Prepare message content
	messageContent := ""
	if m.logMessageBody {
		if messageType == websocket.TextMessage {
			messageContent = string(data)
		} else {
			messageContent = hex.EncodeToString(data)
		}
	}

	// Log the outbound message
	messageLog := model.NewWebSocketMessageLog(
		m.connectionID,
		wsMessageType,
		model.MessageDirectionOutbound,
		int64(len(data)),
		messageContent,
	)
	m.store.Add(messageLog)

	// Call the original WriteMessage method
	return m.Conn.WriteMessage(messageType, data)
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}