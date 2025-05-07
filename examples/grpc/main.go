package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	port             = flag.Int("port", 8080, "HTTP server port")
	grpcPort         = flag.Int("grpc-port", 9090, "gRPC server port")
	enableVisualizer = flag.Bool("viz", true, "Enable request visualizer")
)

// startGRPCServer starts the gRPC server and returns the server and listener.
func startGRPCServer() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var s *grpc.Server

	if *enableVisualizer {
		// Create server with visualizer
		s = govisual.NewGRPCServer(
			govisual.WithGRPC(true),
			govisual.WithGRPCRequestDataLogging(true),
			govisual.WithGRPCResponseDataLogging(true),
			govisual.WithMaxRequests(100),
		)
	} else {
		// Create regular server
		s = grpc.NewServer()
	}

	pb_greeter.RegisterGreeterServiceServer(s, &server{})

	log.Printf("gRPC server listening on port %d", *grpcPort)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return s, lis
}

// handleRoot handles the root HTTP endpoint.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body>
		<h1>GoVisual gRPC Example</h1>
		<p>Visit <a href="/__viz">/__viz</a> to access the request visualizer</p>
		<p>The following gRPC methods are available on port %d:</p>
		<ul>
			<li><code>SayHello</code> - Simple unary RPC</li>
			<li><code>SayHelloStream</code> - Server streaming RPC</li>
			<li><code>CollectHellos</code> - Client streaming RPC</li>
			<li><code>ChatHello</code> - Bidirectional streaming RPC</li>
		</ul>
		<p>Try the sample requests:</p>
		<ul>
			<li><a href="#" class="test-link" data-test="unary">Test Unary RPC</a></li>
			<li><a href="#" class="test-link" data-test="server-stream">Test Server Streaming</a></li>
			<li><a href="#" class="test-link" data-test="client-stream">Test Client Streaming</a></li>
			<li><a href="#" class="test-link" data-test="bidi-stream">Test Bidirectional Streaming</a></li>
			<li><a href="#" class="test-link" data-test="error">Test Error Response</a></li>
			<li><a href="#" class="test-link" data-test="slow">Test Slow Response</a></li>
			<li><a href="#" class="test-link" data-test="metadata">Test with Metadata</a></li>
		</ul>
		<div id="result" style="margin-top: 20px; padding: 10px; border: 1px solid #ccc; display: none;"></div>
		
		<script>
			document.querySelectorAll('.test-link').forEach(link => {
				link.addEventListener('click', async (e) => {
					e.preventDefault();
					const testType = e.target.getAttribute('data-test');
					const resultDiv = document.getElementById('result');
					
					resultDiv.style.display = 'block';
					resultDiv.innerHTML = 'Sending request...';
					
					try {
						const response = await fetch('/test-grpc/' + testType, { method: 'POST' });
						const result = await response.text();
						resultDiv.innerHTML = '<h3>Result:</h3><pre>' + result + '</pre>';
					} catch (error) {
						resultDiv.innerHTML = '<h3>Error:</h3><pre>' + error.message + '</pre>';
					}
				});
			});
		</script>
	</body></html>`, *grpcPort)
}

// makeUnaryRequest makes a unary gRPC request.
func makeUnaryRequest(ctx context.Context, client pb_greeter.GreeterServiceClient, name string) (string, error) {
	resp, err := client.SayHello(ctx, &pb_greeter.HelloRequest{Name: name})
	if err != nil {
		return "", fmt.Errorf("failed to call SayHello: %w", err)
	}
	return resp.GetMessage(), nil
}

// makeServerStreamRequest makes a server streaming gRPC request.
func makeServerStreamRequest(ctx context.Context, client pb_greeter.GreeterServiceClient, name string) ([]string, error) {
	stream, err := client.SayHelloStream(ctx, &pb_greeter.HelloRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to call SayHelloStream: %w", err)
	}

	var responses []string
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return responses, fmt.Errorf("error receiving from stream: %w", err)
		}
		responses = append(responses, resp.GetMessage())
	}

	return responses, nil
}

// makeClientStreamRequest makes a client streaming gRPC request.
func makeClientStreamRequest(ctx context.Context, client pb_greeter.GreeterServiceClient, names []string) (string, error) {
	stream, err := client.CollectHellos(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to call CollectHellos: %w", err)
	}

	for _, name := range names {
		if err := stream.Send(&pb_greeter.HelloRequest{Name: name}); err != nil {
			return "", fmt.Errorf("error sending to stream: %w", err)
		}
		time.Sleep(100 * time.Millisecond) // Space out the requests
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("error receiving response: %w", err)
	}

	return resp.GetMessage(), nil
}

// makeBidiStreamRequest makes a bidirectional streaming gRPC request.
func makeBidiStreamRequest(ctx context.Context, client pb_greeter.GreeterServiceClient, names []string) ([]string, error) {
	stream, err := client.ChatHello(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to call ChatHello: %w", err)
	}

	waitc := make(chan struct{})
	var responses []string

	// Receive responses in a goroutine
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Error receiving from stream: %v", err)
				close(waitc)
				return
			}
			responses = append(responses, resp.GetMessage())
		}
	}()

	// Send requests
	for _, name := range names {
		if err := stream.Send(&pb_greeter.HelloRequest{Name: name}); err != nil {
			return nil, fmt.Errorf("error sending to stream: %w", err)
		}
		time.Sleep(100 * time.Millisecond) // Space out the requests
	}
	stream.CloseSend()

	// Wait for the receiving goroutine to finish
	<-waitc
	return responses, nil
}

// handleTestGRPC handles the test gRPC endpoint.
func handleTestGRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract test type from URL
	testType := strings.TrimPrefix(r.URL.Path, "/test-grpc/")
	if testType == "" {
		testType = "unary" // Default to unary if not specified
	}

	// Connect to the gRPC server
	var conn *grpc.ClientConn
	var err error

	if *enableVisualizer {
		// Connect with visualizer
		conn, err = govisual.DialGRPCWithVisualizer(
			fmt.Sprintf("localhost:%d", *grpcPort),
			govisual.WithGRPC(true),
			govisual.WithGRPCRequestDataLogging(true),
			govisual.WithGRPCResponseDataLogging(true),
		)
	} else {
		// Connect without visualizer
		conn, err = grpc.Dial(
			fmt.Sprintf("localhost:%d", *grpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create client
	client := pb_greeter.NewGreeterServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add metadata to context for metadata test
	if testType == "metadata" {
		md := metadata.New(map[string]string{
			"client-id":    "test-client",
			"request-time": time.Now().Format(time.RFC3339),
			"test-type":    "metadata-test",
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Execute test based on type
	var result string
	switch testType {
	case "unary":
		message, err := makeUnaryRequest(ctx, client, "User")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Unary response: %s", message)

	case "server-stream":
		messages, err := makeServerStreamRequest(ctx, client, "Stream User")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Received %d messages from server stream:\n", len(messages))
		for i, msg := range messages {
			result += fmt.Sprintf("  %d: %s\n", i+1, msg)
		}

	case "client-stream":
		names := []string{"Alice", "Bob", "Charlie", "Dave", "Eve"}
		message, err := makeClientStreamRequest(ctx, client, names)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Client stream response: %s", message)

	case "bidi-stream":
		names := []string{"User 1", "User 2", "User 3", "User 4"}
		messages, err := makeBidiStreamRequest(ctx, client, names)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Received %d messages from bidirectional stream:\n", len(messages))
		for i, msg := range messages {
			result += fmt.Sprintf("  %d: %s\n", i+1, msg)
		}

	case "error":
		// Test error case
		_, err := makeUnaryRequest(ctx, client, "error")
		if err != nil {
			result = fmt.Sprintf("Expected error received: %v", err)
		} else {
			result = "Error test failed: expected an error but got success"
		}

	case "slow":
		// Test slow response
		start := time.Now()
		message, err := makeUnaryRequest(ctx, client, "slow")
		elapsed := time.Since(start)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Slow response received after %v: %s", elapsed, message)

	case "metadata":
		var header, trailer metadata.MD
		resp, err := client.SayHello(
			ctx,
			&pb_greeter.HelloRequest{Name: "Metadata Test"},
			grpc.Header(&header),
			grpc.Trailer(&trailer),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}
		result = fmt.Sprintf("Response with metadata: %s\n\nHeader metadata:\n", resp.GetMessage())
		for k, v := range header {
			result += fmt.Sprintf("  %s: %s\n", k, v)
		}
		result += "\nTrailer metadata:\n"
		for k, v := range trailer {
			result += fmt.Sprintf("  %s: %s\n", k, v)
		}

	default:
		http.Error(w, fmt.Sprintf("Unknown test type: %s", testType), http.StatusBadRequest)
		return
	}

	// Write success response
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(result))
}

func main() {
	flag.Parse()

	// Start gRPC server
	grpcServer, grpcLis := startGRPCServer()

	// Create HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/test-grpc/", handleTestGRPC)

	// Wrap with GoVisual if enabled
	var handler http.Handler = mux
	if *enableVisualizer {
		handler = govisual.Wrap(
			mux,
			govisual.WithRequestBodyLogging(true),
			govisual.WithResponseBodyLogging(true),
		)
	}

	// Create HTTP server with shutdown capability
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("HTTP server started at http://localhost%s", httpServer.Addr)
		if *enableVisualizer {
			log.Printf("Visit http://localhost%s/__viz to see the dashboard", httpServer.Addr)
		}
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Set up signal catching
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Block until signal received
	<-signalChan
	log.Println("Shutdown signal received, exiting...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Gracefully shut down gRPC server
	grpcServer.GracefulStop()
	grpcLis.Close()

	log.Println("Servers shut down successfully")
}
