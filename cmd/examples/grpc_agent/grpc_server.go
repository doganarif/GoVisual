package main

import (
	"context"
	pb_greeter "example/gen/greeter/v1"
	"fmt"
	"time"
)

// Server is used to implement the GreeterServiceServer.
type Server struct {
	pb_greeter.UnimplementedGreeterServiceServer
}

// SayHello implements the SayHello RPC method.
func (s *Server) SayHello(ctx context.Context, req *pb_greeter.HelloRequest) (*pb_greeter.HelloReply, error) {
	return &pb_greeter.HelloReply{
		Message:   "Hello " + req.GetName(),
		Timestamp: time.Now().Unix(),
	}, nil
}

// SayHelloStream implements the server streaming RPC method.
func (s *Server) SayHelloStream(req *pb_greeter.HelloRequest, stream pb_greeter.GreeterService_SayHelloStreamServer) error {
	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb_greeter.HelloReply{
			Message:   fmt.Sprintf("Hello %s! (response #%d)", req.GetName(), i+1),
			Timestamp: time.Now().Unix(),
		}); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// CollectHellos implements the client streaming RPC method.
func (s *Server) CollectHellos(stream pb_greeter.GreeterService_CollectHellosServer) error {
	var names []string

	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}
		names = append(names, req.GetName())
	}

	var message string
	if len(names) == 0 {
		message = "Hello to nobody!"
	} else if len(names) == 1 {
		message = fmt.Sprintf("Hello to %s!", names[0])
	} else {
		message = fmt.Sprintf("Hello to %d people!", len(names))
	}

	return stream.SendAndClose(&pb_greeter.HelloReply{
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// ChatHello implements the bidirectional streaming RPC method.
func (s *Server) ChatHello(stream pb_greeter.GreeterService_ChatHelloServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil
		}

		if err := stream.Send(&pb_greeter.HelloReply{
			Message:   "Hello " + req.GetName() + "!",
			Timestamp: time.Now().Unix(),
		}); err != nil {
			return err
		}
	}
}
