package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb_greeter "example/gen/greeter/v1"
)

// server is used to implement the GreeterServer.
type server struct {
	pb_greeter.UnimplementedGreeterServiceServer
}

// SayHello implements the gRPC SayHello method.
func (s *server) SayHello(ctx context.Context, req *pb_greeter.HelloRequest) (*pb_greeter.HelloReply, error) {
	log.Printf("Received: %v", req.GetName())

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	return &pb_greeter.HelloReply{Message: "Hello " + req.GetName()}, nil
}

// SayHelloStream implements a server streaming RPC method.
func (s *server) SayHelloStream(req *pb_greeter.HelloRequest, stream pb_greeter.GreeterService_SayHelloStreamServer) error {
	log.Printf("Received stream request from: %v", req.GetName())

	// Send multiple responses
	for i := 0; i < 5; i++ {
		// Simulate processing time between messages
		time.Sleep(200 * time.Millisecond)

		// Send response
		if err := stream.Send(&pb_greeter.HelloReply{
			Message: fmt.Sprintf("Hello %s! (response #%d)", req.GetName(), i+1),
		}); err != nil {
			return err
		}
	}

	return nil
}

// CollectHellos implements a client streaming RPC method.
func (s *server) CollectHellos(stream pb_greeter.GreeterService_CollectHellosServer) error {
	var names []string

	// Collect client messages
	for {
		req, err := stream.Recv()
		if err != nil {
			// End of stream
			break
		}

		log.Printf("Received in client stream: %v", req.GetName())
		names = append(names, req.GetName())

		// Simulate processing
		time.Sleep(100 * time.Millisecond)
	}

	// Send response after collecting all messages
	response := fmt.Sprintf("Hello to: %v", names)
	return stream.SendAndClose(&pb_greeter.HelloReply{Message: response})
}

// ChatHello implements a bidirectional streaming RPC method.
func (s *server) ChatHello(stream pb_greeter.GreeterService_ChatHelloServer) error {
	// Process messages as they come in and respond immediately
	for {
		req, err := stream.Recv()
		if err != nil {
			// End of stream
			return nil
		}

		log.Printf("Received in bidi stream: %v", req.GetName())

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		// Send response
		if err := stream.Send(&pb_greeter.HelloReply{
			Message: fmt.Sprintf("Hello %s!", req.GetName()),
		}); err != nil {
			return err
		}
	}
}
