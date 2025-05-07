package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb_greeter "example/gen/greeter/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server is used to implement the GreeterServiceServer.
type Server struct {
	pb_greeter.UnimplementedGreeterServiceServer
}

// SayHello implements a unary RPC method (single request, single response).
// This is the simplest type of gRPC call.
func (s *Server) SayHello(ctx context.Context, req *pb_greeter.HelloRequest) (*pb_greeter.HelloReply, error) {
	log.Printf("Unary call - SayHello called with name: %v", req.GetName())

	// Extract any metadata from the request
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, values := range md {
			log.Printf("  Metadata: %s = %v", key, values)
		}
	}

	// Send header metadata to the client
	header := metadata.New(map[string]string{
		"server-time": time.Now().Format(time.RFC3339),
		"server-name": "govisual-example",
	})
	if err := grpc.SendHeader(ctx, header); err != nil {
		log.Printf("Failed to send header: %v", err)
	}

	// Special case for testing errors
	if req.GetName() == "error" {
		log.Printf("Returning error for test case")
		return nil, status.Errorf(codes.InvalidArgument, "Name 'error' triggers a deliberate error")
	}

	// Special case for testing slow responses
	if req.GetName() == "slow" {
		log.Printf("Slow response test - waiting 1 second")
		time.Sleep(1 * time.Second)
	}

	// Create and return the response
	response := &pb_greeter.HelloReply{
		Message:   "Hello " + req.GetName(),
		Timestamp: time.Now().Unix(),
	}

	return response, nil
}

// SayHelloStream implements a server streaming RPC method.
// The server sends multiple responses for a single client request.
func (s *Server) SayHelloStream(req *pb_greeter.HelloRequest, stream pb_greeter.GreeterService_SayHelloStreamServer) error {
	log.Printf("Server streaming - SayHelloStream called with name: %v", req.GetName())

	// Send response header metadata
	header := metadata.New(map[string]string{
		"stream-start":   time.Now().Format(time.RFC3339),
		"response-count": "5",
	})
	if err := stream.SendHeader(header); err != nil {
		log.Printf("Failed to send header: %v", err)
	}

	// Stream 5 responses back to the client
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)

		response := &pb_greeter.HelloReply{
			Message:   fmt.Sprintf("Hello %s! (response #%d)", req.GetName(), i+1),
			Timestamp: time.Now().Unix(),
		}

		if err := stream.Send(response); err != nil {
			log.Printf("Error sending stream response: %v", err)
			return status.Errorf(codes.Internal, "Failed to send response: %v", err)
		}

		log.Printf("Sent streaming response #%d", i+1)
	}

	// Send trailer metadata
	trailer := metadata.New(map[string]string{
		"stream-end": time.Now().Format(time.RFC3339),
		"status":     "completed",
	})
	stream.SetTrailer(trailer)

	return nil
}

// CollectHellos implements a client streaming RPC method.
// The client sends multiple requests, and the server sends a single response.
func (s *Server) CollectHellos(stream pb_greeter.GreeterService_CollectHellosServer) error {
	log.Printf("Client streaming - CollectHellos started")

	var names []string
	count := 0

	// Send header metadata
	header := metadata.New(map[string]string{
		"stream-type": "client-streaming",
		"start-time":  time.Now().Format(time.RFC3339),
	})
	if err := stream.SendHeader(header); err != nil {
		log.Printf("Failed to send header: %v", err)
	}

	// Receive client messages until the client closes the stream
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("End of client stream: %v", err)
			break
		}

		count++
		log.Printf("Received client message #%d: name=%s", count, req.GetName())
		names = append(names, req.GetName())

		// Small delay to simulate processing
		time.Sleep(100 * time.Millisecond)
	}

	// Prepare message for names
	var message string
	if len(names) == 0 {
		message = "Hello to nobody!"
	} else if len(names) == 1 {
		message = fmt.Sprintf("Hello to %s!", names[0])
	} else {
		message = fmt.Sprintf("Hello to %d people: %v!", len(names), names)
	}

	// Set trailer metadata
	trailer := metadata.New(map[string]string{
		"message-count": fmt.Sprintf("%d", count),
		"end-time":      time.Now().Format(time.RFC3339),
	})
	stream.SetTrailer(trailer)

	// Send a single response with all collected names
	return stream.SendAndClose(&pb_greeter.HelloReply{
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// ChatHello implements a bidirectional streaming RPC method.
// Both client and server can send messages independently.
func (s *Server) ChatHello(stream pb_greeter.GreeterService_ChatHelloServer) error {
	log.Printf("Bidirectional streaming - ChatHello started")

	// Send header metadata
	header := metadata.New(map[string]string{
		"stream-type": "bidirectional",
		"start-time":  time.Now().Format(time.RFC3339),
	})
	if err := stream.SendHeader(header); err != nil {
		log.Printf("Failed to send header: %v", err)
	}

	messageCount := 0

	// Process messages as they arrive
	for {
		// Wait for the next message from the client
		req, err := stream.Recv()
		if err != nil {
			log.Printf("End of bidirectional stream: %v", err)

			// Set trailer metadata before returning
			trailer := metadata.New(map[string]string{
				"message-count": fmt.Sprintf("%d", messageCount),
				"end-time":      time.Now().Format(time.RFC3339),
			})
			stream.SetTrailer(trailer)

			return nil
		}

		messageCount++
		log.Printf("Received bidi message #%d: name=%s", messageCount, req.GetName())

		// Create and send a response
		response := &pb_greeter.HelloReply{
			Message:   fmt.Sprintf("Hello %s! (reply #%d)", req.GetName(), messageCount),
			Timestamp: time.Now().Unix(),
		}

		if err := stream.Send(response); err != nil {
			log.Printf("Error sending bidi response: %v", err)
			return status.Errorf(codes.Internal, "Failed to send response: %v", err)
		}

		log.Printf("Sent bidi response #%d", messageCount)
	}
}
