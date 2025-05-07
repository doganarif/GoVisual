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
	"syscall"
	"time"

	pb_greeter "example/gen/greeter/v1"

	"github.com/doganarif/govisual"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port     = flag.Int("port", 8080, "HTTP server port (for dashboard)")
	grpcPort = flag.Int("grpc-port", 9090, "gRPC server port")
)

// startGRPCServer starts the gRPC server with GoVisual instrumentation.
func startGRPCServer() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create server with GoVisual's NewGRPCServer
	server := govisual.NewGRPCServer(
		govisual.WithGRPC(true),
		govisual.WithGRPCRequestDataLogging(true),
		govisual.WithGRPCResponseDataLogging(true),
		govisual.WithMaxRequests(100),
	)

	// Register the GreeterService
	pb_greeter.RegisterGreeterServiceServer(server, &Server{})

	log.Printf("gRPC server listening on port %d with visualization", *grpcPort)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return server, lis
}

// startDashboardServer starts the HTTP server for the dashboard.
func startDashboardServer() *http.Server {
	// Create a simple HTTP mux
	mux := http.NewServeMux()

	// Add a basic homepage
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf(`
			<html>
			<head><title>GoVisual gRPC Example</title></head>
			<body>
				<h1>GoVisual gRPC Example</h1>
				<p>Visit <a href="/__viz">/__viz</a> to access the request visualizer.</p>
				<p>gRPC server is running on port %d.</p>
				<p>Use grpcui or other gRPC clients to interact with the server.</p>
			</body>
			</html>
		`, *grpcPort)))
	})

	// Wrap with GoVisual
	handler := govisual.Wrap(
		mux,
		govisual.WithMaxRequests(100),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	// Create and start the HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler,
	}

	go func() {
		log.Printf("Dashboard server started at http://localhost:%d", *port)
		log.Printf("Visit http://localhost:%d/__viz to see the dashboard", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	return server
}

// makeTestRequest makes a test gRPC request.
func makeTestRequest() {
	// Wait for servers to start
	time.Sleep(500 * time.Millisecond)

	// Connect to the gRPC server
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", *grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Failed to connect for test request: %v", err)
		return
	}
	defer conn.Close()

	// Create client
	client := pb_greeter.NewGreeterServiceClient(conn)

	// Make request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Sending test gRPC request - check the dashboard to see it")
	resp, err := client.SayHello(ctx, &pb_greeter.HelloRequest{Name: "Test"})
	if err != nil {
		log.Printf("Test request failed: %v", err)
		return
	}
	log.Printf("Test response received: %s", resp.GetMessage())
}

func main() {
	flag.Parse()

	// Start the gRPC server
	grpcServer, grpcLis := startGRPCServer()

	// Start the dashboard server
	httpServer := startDashboardServer()

	// Make a test request
	go makeTestRequest()

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Wait for termination signal
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
	grpcLis.Close()

	log.Println("Servers shut down successfully")
}
