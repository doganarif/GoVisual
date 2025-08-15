package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/doganarif/govisual"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader with CORS allowing all origins for demo purposes
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Handle messages
	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		log.Printf("Received message: %s", message)

		// Echo the message back to client
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Failed to write message: %v", err)
			break
		}
	}
}

// handleChatRoom handles a chat room WebSocket connection
func handleChatRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get room name from query parameter
	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		roomName = "general"
	}

	log.Printf("User joined room: %s", roomName)

	// Send welcome message
	welcomeMsg := fmt.Sprintf("Welcome to room: %s", roomName)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg)); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Handle chat messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process the message (in a real app, you'd broadcast to all users in the room)
		chatMessage := fmt.Sprintf("[%s] %s", roomName, string(message))
		log.Printf("Chat message: %s", chatMessage)

		// Echo back with room prefix
		if err := conn.WriteMessage(messageType, []byte(chatMessage)); err != nil {
			log.Printf("Failed to echo chat message: %v", err)
			break
		}
	}
}

func main() {
	// Create a simple HTTP handler for the home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Monitoring Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        #messages { height: 300px; overflow-y: scroll; border: 1px solid #ccc; padding: 10px; margin: 10px 0; }
        input { padding: 10px; margin: 5px; width: 200px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>WebSocket Monitoring Example</h1>
        
        <div class="section">
            <h2>Echo Server Test</h2>
            <button onclick="connectEcho()">Connect to Echo Server</button>
            <button onclick="disconnectEcho()">Disconnect</button>
            <br>
            <input type="text" id="echoMessage" placeholder="Type a message...">
            <button onclick="sendEchoMessage()">Send Message</button>
            <div id="echoMessages"></div>
        </div>

        <div class="section">
            <h2>Chat Room Test</h2>
            <input type="text" id="roomName" placeholder="Room name" value="general">
            <button onclick="connectChat()">Connect to Chat</button>
            <button onclick="disconnectChat()">Disconnect</button>
            <br>
            <input type="text" id="chatMessage" placeholder="Type a chat message...">
            <button onclick="sendChatMessage()">Send Message</button>
            <div id="chatMessages"></div>
        </div>

        <div class="section">
            <h2>Monitoring Dashboard</h2>
            <p>Visit <a href="/__viz/" target="_blank">/__viz/</a> to see WebSocket monitoring dashboard</p>
            <p>You can filter by:</p>
            <ul>
                <li>Connection Type: WebSocket vs HTTP</li>
                <li>Message Type: text, binary, ping, pong, close</li>
                <li>Message Direction: inbound vs outbound</li>
                <li>Connection ID: track specific WebSocket connections</li>
            </ul>
        </div>
    </div>

    <script>
        let echoWs = null;
        let chatWs = null;

        function connectEcho() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            echoWs = new WebSocket(protocol + '//' + window.location.host + '/echo');
            
            echoWs.onopen = function() {
                addMessage('echoMessages', 'Connected to echo server');
            };
            
            echoWs.onmessage = function(event) {
                addMessage('echoMessages', 'Received: ' + event.data);
            };
            
            echoWs.onclose = function() {
                addMessage('echoMessages', 'Disconnected from echo server');
            };
        }

        function disconnectEcho() {
            if (echoWs) {
                echoWs.close();
            }
        }

        function sendEchoMessage() {
            const message = document.getElementById('echoMessage').value;
            if (echoWs && message) {
                echoWs.send(message);
                addMessage('echoMessages', 'Sent: ' + message);
                document.getElementById('echoMessage').value = '';
            }
        }

        function connectChat() {
            const roomName = document.getElementById('roomName').value || 'general';
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            chatWs = new WebSocket(protocol + '//' + window.location.host + '/chat?room=' + roomName);
            
            chatWs.onopen = function() {
                addMessage('chatMessages', 'Connected to chat room: ' + roomName);
            };
            
            chatWs.onmessage = function(event) {
                addMessage('chatMessages', event.data);
            };
            
            chatWs.onclose = function() {
                addMessage('chatMessages', 'Disconnected from chat');
            };
        }

        function disconnectChat() {
            if (chatWs) {
                chatWs.close();
            }
        }

        function sendChatMessage() {
            const message = document.getElementById('chatMessage').value;
            if (chatWs && message) {
                chatWs.send(message);
                document.getElementById('chatMessage').value = '';
            }
        }

        function addMessage(elementId, message) {
            const div = document.getElementById(elementId);
            const p = document.createElement('p');
            p.textContent = new Date().toLocaleTimeString() + ': ' + message;
            div.appendChild(p);
            div.scrollTop = div.scrollHeight;
        }

        // Allow Enter key to send messages
        document.getElementById('echoMessage').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') sendEchoMessage();
        });
        
        document.getElementById('chatMessage').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') sendChatMessage();
        });
    </script>
</body>
</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// Wrap WebSocket handlers with monitoring
	wrappedEcho := govisual.WrapWebSocket(
		handleWebSocket,
		govisual.WithDashboardPath("/__viz"),
		govisual.WithRequestBodyLogging(true), // Log WebSocket message content
		govisual.WithMaxRequests(1000),
	)

	wrappedChat := govisual.WrapWebSocket(
		handleChatRoom,
		govisual.WithDashboardPath("/__viz"),
		govisual.WithRequestBodyLogging(true),
		govisual.WithMaxRequests(1000),
	)

	// Register WebSocket handlers
	http.Handle("/echo", wrappedEcho)
	http.Handle("/chat", wrappedChat)

	// Also show HTTP monitoring by wrapping a simple API endpoint
	apiHandler := govisual.Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "Hello from HTTP API!", "timestamp": "` + 
				fmt.Sprintf("%d", 1234567890) + `"}`))
		}),
		govisual.WithDashboardPath("/__viz"),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)
	http.Handle("/api/hello", apiHandler)

	fmt.Println("🚀 Server starting on :8080")
	fmt.Println("📊 Dashboard available at: http://localhost:8080/__viz/")
	fmt.Println("🔌 WebSocket Echo at: ws://localhost:8080/echo")
	fmt.Println("💬 WebSocket Chat at: ws://localhost:8080/chat?room=general")
	fmt.Println("🌐 HTTP API at: http://localhost:8080/api/hello")
	fmt.Println()
	fmt.Println("Visit http://localhost:8080/ for interactive WebSocket test client")

	log.Fatal(http.ListenAndServe(":8080", nil))
}