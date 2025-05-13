package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb_greeter "example/gen/greeter/v1"

	"github.com/doganarif/govisual"
	"github.com/doganarif/govisual/internal/store"
	"github.com/doganarif/govisual/pkg/agent"
	"github.com/doganarif/govisual/pkg/server"

	"github.com/doganarif/govisual/pkg/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port      = flag.Int("port", 8080, "HTTP server port (for dashboard)")
	grpcPort  = flag.Int("grpc-port", 9090, "gRPC server port")
	agentMode = flag.String("agent-mode", "store", "Agent mode: store, nats, http")
	natsURL   = flag.String("nats-url", "nats://localhost:4222", "NATS server URL")
	httpURL   = flag.String("http-url", "http://localhost:8080/api/agent/logs", "HTTP endpoint URL")
)

func main() {
	flag.Parse()

	// Create a shared store for visualization
	sharedStore, err := govisual.NewStore(
		govisual.WithMaxRequests(100),
		govisual.WithMemoryStorage(),
	)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Create transport based on agent mode
	var transportObj transport.Transport

	switch *agentMode {
	case "store":
		log.Println("Using shared store transport")
		transportObj = transport.NewStoreTransport(sharedStore)
	case "nats":
		log.Printf("Using NATS transport with server: %s", *natsURL)
		transportObj, err = transport.NewNATSTransport(*natsURL)
		if err != nil {
			log.Fatalf("Failed to create NATS transport: %v", err)
		}
	case "http":
		log.Printf("Using HTTP transport with endpoint: %s", *httpURL)
		transportObj = transport.NewHTTPTransport(*httpURL,
			transport.WithTimeout(5*time.Second),
			transport.WithMaxRetries(3),
		)
	default:
		log.Fatalf("Unknown agent mode: %s", *agentMode)
	}

	// Create gRPC agent
	grpcAgent := agent.NewGRPCAgent(transportObj,
		agent.WithGRPCRequestDataLogging(true),
		agent.WithGRPCResponseDataLogging(true),
		agent.WithBatchingEnabled(true),
		agent.WithBatchSize(5),
		agent.WithBatchInterval(1*time.Second),
	)

	// Start the gRPC server with the agent
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := server.NewGRPCServer(grpcAgent)
	pb_greeter.RegisterGreeterServiceServer(grpcServer, &Server{})

	log.Printf("gRPC server listening on port %d with visualization", *grpcPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Start the HTTP dashboard server
	mux := http.NewServeMux()

	// Add a simple homepage
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf(`
			<html>
			<head>
				<title>GoVisual gRPC Agent Example</title>
				<style>
					body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
					h1 { color: #333; }
					.container { max-width: 800px; margin: 0 auto; }
					.card { background: #f9f9f9; border-radius: 5px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
					.info { background: #e8f4f8; padding: 15px; border-radius: 5px; margin-bottom: 15px; }
					a { color: #0066cc; text-decoration: none; }
					a:hover { text-decoration: underline; }
					.btn { display: inline-block; background: #0066cc; color: white; padding: 10px 15px; border-radius: 5px; text-decoration: none; margin-top: 10px; }
					.btn:hover { background: #0055aa; }
					code { background: #f0f0f0; padding: 2px 5px; border-radius: 3px; font-family: monospace; }
				</style>
			</head>
			<body>
				<div class="container">
					<h1>GoVisual gRPC Agent Example</h1>
					<div class="card">
						<h2>Dashboard</h2>
						<p>Visit <a href="/__viz">/__viz</a> to access the request visualizer dashboard.</p>
					</div>
					<div class="card">
						<h2>Configuration</h2>
						<div class="info">
							<strong>gRPC Server:</strong> localhost:%d<br>
							<strong>Agent Mode:</strong> %s<br>
							%s
						</div>
					</div>
					<div class="card">
						<h2>Test the gRPC Service</h2>
						<p>An initial test request has been made automatically. You can make additional requests using a gRPC client.</p>
						<p>The service provides the following methods:</p>
						<ul>
							<li><code>SayHello</code> - Unary RPC</li>
							<li><code>SayHelloStream</code> - Server streaming RPC</li>
							<li><code>CollectHellos</code> - Client streaming RPC</li>
							<li><code>ChatHello</code> - Bidirectional streaming RPC</li>
						</ul>
						<p>You can use a tool like <a href="https://github.com/fullstorydev/grpcui" target="_blank">grpcui</a> or <a href="https://github.com/fullstorydev/grpcurl" target="_blank">grpcurl</a> to test these methods.</p>
					</div>
				</div>
			</body>
			</html>
		`, *grpcPort, *agentMode, getExtraInfo())))
	})

	// Add API endpoints for agent communication
	if *agentMode != "store" {
		// Type assert to get the internal store implementation
		internalStore, ok := sharedStore.(store.Store)
		if !ok {
			internalStore = store.NewInMemoryStore(100)
			log.Println("Warning: Failed to convert shared store, created new in-memory store for agent API")
		}

		// Create and register agent API handler
		agentAPI := server.NewAgentAPI(internalStore)
		agentAPI.RegisterHandlers(mux)
	}

	// Wrap with GoVisual for dashboard
	handler := govisual.Wrap(
		mux,
		govisual.WithMaxRequests(100),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithSharedStore(sharedStore),
	)

	// Start NATS handler if using NATS transport
	var natsHandler *server.NATSHandler
	if *agentMode == "nats" {
		// Type assert to get the internal store implementation
		internalStore, ok := sharedStore.(store.Store)
		if !ok {
			internalStore = store.NewInMemoryStore(100)
			log.Println("Warning: Failed to convert shared store, created new in-memory store for NATS handler")
		}

		var err error
		natsHandler, err = server.NewNATSHandler(internalStore, *natsURL)
		if err != nil {
			log.Fatalf("Failed to create NATS handler: %v", err)
		}

		if err := natsHandler.Start(); err != nil {
			log.Fatalf("Failed to start NATS handler: %v", err)
		}
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler,
	}

	log.Printf("Dashboard server started at http://localhost:%d", *port)
	log.Printf("Visit http://localhost:%d/__viz to see the dashboard", *port)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Make test requests to show different gRPC method types
	go func() {
		time.Sleep(500 * time.Millisecond)

		// Create gRPC client
		conn, err := grpc.Dial(
			fmt.Sprintf("localhost:%d", *grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Printf("Failed to connect: %v", err)
			return
		}
		defer conn.Close()

		client := pb_greeter.NewGreeterServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test 1: Unary RPC
		log.Println("Testing unary RPC (SayHello)")
		resp, err := client.SayHello(ctx, &pb_greeter.HelloRequest{
			Name:    "Agent Test",
			Message: "This is a test message",
		})
		if err != nil {
			log.Printf("SayHello failed: %v", err)
		} else {
			log.Printf("SayHello response: %s (timestamp: %d)", resp.GetMessage(), resp.GetTimestamp())
		}

		// Test 2: Server streaming RPC
		log.Println("Testing server streaming RPC (SayHelloStream)")
		stream, err := client.SayHelloStream(ctx, &pb_greeter.HelloRequest{
			Name:    "Stream Test",
			Message: "Testing server streaming",
		})
		if err != nil {
			log.Printf("SayHelloStream failed: %v", err)
		} else {
			for {
				resp, err := stream.Recv()
				if err != nil {
					break
				}
				log.Printf("Stream response: %s (timestamp: %d)", resp.GetMessage(), resp.GetTimestamp())
			}
		}

		// Test 3: Client streaming RPC
		log.Println("Testing client streaming RPC (CollectHellos)")
		clientStream, err := client.CollectHellos(ctx)
		if err != nil {
			log.Printf("CollectHellos failed: %v", err)
		} else {
			// Send multiple messages
			for i := 1; i <= 3; i++ {
				name := fmt.Sprintf("Person %d", i)
				if err := clientStream.Send(&pb_greeter.HelloRequest{
					Name:    name,
					Message: fmt.Sprintf("Message from %s", name),
				}); err != nil {
					log.Printf("Error sending client stream message: %v", err)
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			// Close and receive response
			resp, err := clientStream.CloseAndRecv()
			if err != nil {
				log.Printf("Error closing client stream: %v", err)
			} else {
				log.Printf("Client stream response: %s (timestamp: %d)", resp.GetMessage(), resp.GetTimestamp())
			}
		}

		// Test 4: Bidirectional streaming RPC
		log.Println("Testing bidirectional streaming RPC (ChatHello)")
		bidiStream, err := client.ChatHello(ctx)
		if err != nil {
			log.Printf("ChatHello failed: %v", err)
		} else {
			// Send and receive in goroutines
			done := make(chan bool)

			// Receiving goroutine
			go func() {
				for {
					resp, err := bidiStream.Recv()
					if err != nil {
						break
					}
					log.Printf("Bidi response: %s (timestamp: %d)", resp.GetMessage(), resp.GetTimestamp())
				}
				done <- true
			}()

			// Send messages
			for i := 1; i <= 3; i++ {
				name := fmt.Sprintf("ChatPerson %d", i)
				if err := bidiStream.Send(&pb_greeter.HelloRequest{
					Name:    name,
					Message: fmt.Sprintf("Bidi message from %s", name),
				}); err != nil {
					log.Printf("Error sending bidi message: %v", err)
					break
				}
				time.Sleep(200 * time.Millisecond)
			}

			// Close sending
			bidiStream.CloseSend()
			<-done
		}

		log.Println("All gRPC tests completed")
	}()

	// Wait for termination signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	lis.Close()

	// Stop NATS handler if running
	if natsHandler != nil {
		if err := natsHandler.Stop(); err != nil {
			log.Printf("NATS handler stop error: %v", err)
		}
	}

	// Close the agent
	if err := grpcAgent.Close(); err != nil {
		log.Printf("Agent closure error: %v", err)
	}

	// Close the transport
	if err := transportObj.Close(); err != nil {
		log.Printf("Transport closure error: %v", err)
	}

	// Close the store
	if err := sharedStore.Close(); err != nil {
		log.Printf("Store closure error: %v", err)
	}

	log.Println("Servers shut down successfully")
}

// getExtraInfo returns extra information for the homepage based on the agent mode
func getExtraInfo() string {
	var info strings.Builder

	switch *agentMode {
	case "store":
		info.WriteString("<strong>Transport:</strong> In-memory shared store (direct)")
	case "nats":
		info.WriteString(fmt.Sprintf("<strong>Transport:</strong> NATS messaging via %s", *natsURL))
	case "http":
		info.WriteString(fmt.Sprintf("<strong>Transport:</strong> HTTP via %s", *httpURL))
	}

	return info.String()
}
