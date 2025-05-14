package main

import (
	"context"
	pb_greeter "example/gen/greeter/v1"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/doganarif/govisual"
	"github.com/doganarif/govisual/pkg/agent"
	"github.com/doganarif/govisual/pkg/server"
	"github.com/doganarif/govisual/pkg/store"
	"github.com/doganarif/govisual/pkg/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port      = flag.Int("port", 8080, "HTTP server port (for dashboard)")
	grpcPort  = flag.Int("grpc-port", 9090, "gRPC server port")
	agentMode = flag.String("agent-mode", "store", "Agent mode: store, nats, http")
	natsURL   = flag.String("nats-url", "nats://localhost:4222", "NATS server URL. Only used with agent-mode 'nats'")
	httpURL   = flag.String("http-url", "http://localhost:8080/api/agent/logs", "HTTP endpoint URL. Only used with agent-mode 'http'")
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Printf("failed to run: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sharedStore := store.NewInMemoryStore(100)

	var transportObj transport.Transport
	var err error
	switch *agentMode {
	case "store":
		log.Info("Using shared store transport")
		transportObj = transport.NewStoreTransport(sharedStore)
	case "nats":
		log.Info("Using NATS transport", slog.String("url", *natsURL))
		transportObj, err = transport.NewNATSTransport(*natsURL)
		if err != nil {
			return fmt.Errorf("creating NATS transport: %w", err)
		}
	case "http":
		log.Info("Using HTTP transport", slog.String("url", *httpURL))
		transportObj = transport.NewHTTPTransport(*httpURL,
			transport.WithTimeout(5*time.Second),
			transport.WithMaxRetries(3),
		)
	default:
		return fmt.Errorf("unknown agent mode %q", *agentMode)
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
		return fmt.Errorf("listening to gRPC: %w", err)
	}

	grpcServer := server.NewGRPCServer(grpcAgent)
	pb_greeter.RegisterGreeterServiceServer(grpcServer, &Server{})

	log.Info("Starting gRPC server with visualisation", slog.Int("port", *grpcPort))

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Error("failed to serve gRPC", slog.Any("err", err))
			os.Exit(1)
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
		// Create and register agent API handler
		agentAPI := server.NewAgentAPI(sharedStore)
		agentAPI.RegisterHandlers(mux)
	}

	// Wrap with GoVisual for dashboard
	handler := govisual.Wrap(mux,
		govisual.WithMaxRequests(100),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithSharedStore(sharedStore),
	)

	// Start NATS handler if using NATS transport
	var natsHandler *server.NATSHandler
	if *agentMode == "nats" {
		natsHandler, err = server.NewNATSHandler(sharedStore, *natsURL)
		if err != nil {
			return fmt.Errorf("creating NATS handler: %w", err)
		}

		err = natsHandler.Start()
		if err != nil {
			return fmt.Errorf("starting NATS handler: %w", err)
		}
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler,
	}

	log.Info("Started dashboard server", slog.Int("port", *port), slog.String("dashboard_addr", fmt.Sprintf("http://localhost:%d/__viz", *port)))

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error("failed to serve HTTP", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	// Make test requests to show different gRPC method types
	go func() {
		time.Sleep(500 * time.Millisecond)

		// Create gRPC client
		conn, err := grpc.NewClient(
			fmt.Sprintf("localhost:%d", *grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Error("Failed to connect: %v", err)
			return
		}
		defer conn.Close()

		client := pb_greeter.NewGreeterServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test 1: Unary RPC
		log.Info("Testing unary RPC (SayHello)")
		resp, err := client.SayHello(ctx, &pb_greeter.HelloRequest{
			Name:    "Agent Test",
			Message: "This is a test message",
		})
		if err != nil {
			log.Error("SayHello failed", slog.Any("err", err))
		} else {
			log.Info("Received unary RPC response", slog.String("msg", resp.GetMessage()), slog.Int64("timestamp", resp.GetTimestamp()))
		}

		// Test 2: Server streaming RPC
		log.Info("Testing server streaming RPC (SayHelloStream)")
		stream, err := client.SayHelloStream(ctx, &pb_greeter.HelloRequest{
			Name:    "Stream Test",
			Message: "Testing server streaming",
		})
		if err != nil {
			log.Error("SayHelloStream failed", slog.Any("err", err))
		} else {
			for {
				resp, err := stream.Recv()
				if err != nil {
					break
				}
				log.Info("Received server stream response", slog.String("msg", resp.GetMessage()), slog.Int64("timestamp", resp.GetTimestamp()))
			}
		}

		// Test 3: Client streaming RPC
		log.Info("Testing client streaming RPC (CollectHellos)")
		clientStream, err := client.CollectHellos(ctx)
		if err != nil {
			log.Error("CollectHellos failed", slog.Any("err", err))
		} else {
			// Send multiple messages
			for i := 1; i <= 3; i++ {
				name := fmt.Sprintf("Person %d", i)
				if err := clientStream.Send(&pb_greeter.HelloRequest{
					Name:    name,
					Message: fmt.Sprintf("Message from %s", name),
				}); err != nil {
					log.Error("failed to send client stream message", slog.Any("err", err))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			// Close and receive response
			resp, err := clientStream.CloseAndRecv()
			if err != nil {
				log.Error("failed to close client stream", slog.Any("err", err))
			} else {
				log.Info("Received client stream response", slog.String("msg", resp.GetMessage()), slog.Int64("timestamp", resp.GetTimestamp()))
			}
		}

		// Test 4: Bidirectional streaming RPC
		log.Info("Testing bidirectional streaming RPC (ChatHello)")
		bidiStream, err := client.ChatHello(ctx)
		if err != nil {
			log.Error("ChatHello failed", slog.Any("err", err))
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
					log.Info("Received Bidi response", slog.String("msg", resp.GetMessage()), slog.Int64("timestamp", resp.GetTimestamp()))
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
					log.Error("failed to send bidi message: %v", err)
					break
				}
				time.Sleep(200 * time.Millisecond)
			}

			// Close sending
			bidiStream.CloseSend()
			<-done
		}

		log.Info("All gRPC tests completed")
	}()

	// Wait for termination signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	log.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Info("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	lis.Close()

	if natsHandler != nil {
		err = natsHandler.Stop()
		if err != nil {
			log.Error("failed to stop NATS handler", slog.Any("err", err))
		}
	}

	err = grpcAgent.Close()
	if err != nil {
		log.Error("failed to close agent", slog.Any("err", err))
	}

	err = transportObj.Close()
	if err != nil {
		log.Error("failed to close transport", slog.Any("err", err))
	}

	err = sharedStore.Close()
	if err != nil {
		log.Error("failed to close store", slog.Any("err", err))
	}

	log.Info("Servers shut down successfully")
	return nil
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
