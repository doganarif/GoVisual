package server

import (
	"context"
	"fmt"

	"github.com/doganarif/govisual/pkg/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewGRPCServer creates a new gRPC server with the provided agent.
func NewGRPCServer(agent *agent.GRPCAgent) *grpc.Server {
	// Create unary and stream interceptors
	unaryInterceptor := agent.UnaryServerInterceptor()
	streamInterceptor := agent.StreamServerInterceptor()

	// Create server with interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptor),
		grpc.ChainStreamInterceptor(streamInterceptor),
	)

	return server
}

// gRPCClient is a gRPC client with integrated GoVisual agent.
type gRPCClient struct {
	agent *agent.GRPCAgent
	conn  *grpc.ClientConn
}

// NewGRPCClient creates a new gRPC client with the provided agent.
func NewGRPCClient(target string, agent *agent.GRPCAgent, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Create unary and stream interceptors
	unaryInterceptor := agent.UnaryClientInterceptor()
	streamInterceptor := agent.StreamClientInterceptor()

	// Add interceptors to dial options
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(unaryInterceptor),
		grpc.WithChainStreamInterceptor(streamInterceptor),
	)

	// Create connection
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return conn, nil
}

// CloseConnection closes the gRPC client connection.
func CloseConnection(conn *grpc.ClientConn) error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// ErrorInterceptor is a simple interceptor that returns an error for testing purposes.
func ErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return nil, status.Errorf(codes.Internal, "error interceptor")
}
