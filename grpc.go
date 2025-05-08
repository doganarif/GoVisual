package govisual

import (
	"context"
	"log"

	internal_grpc "github.com/doganarif/govisual/internal/grpc"
	"github.com/doganarif/govisual/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCConfig contains configuration options for gRPC support.
type GRPCConfig struct {
	// EnableGRPC determines whether gRPC interceptors are enabled.
	EnableGRPC bool

	// LogRequestData determines whether request message data is logged.
	LogRequestData bool

	// LogResponseData determines whether response message data is logged.
	LogResponseData bool

	// IgnoreMethods is a list of method patterns to ignore.
	IgnoreMethods []string
}

// WithGRPC enables or disables gRPC interceptors.
func WithGRPC(enabled bool) Option {
	return func(c *Config) {
		c.GRPC.EnableGRPC = enabled
	}
}

// WithGRPCRequestDataLogging enables or disables logging of gRPC request message data.
func WithGRPCRequestDataLogging(enabled bool) Option {
	return func(c *Config) {
		c.GRPC.LogRequestData = enabled
	}
}

// WithGRPCResponseDataLogging enables or disables logging of gRPC response message data.
func WithGRPCResponseDataLogging(enabled bool) Option {
	return func(c *Config) {
		c.GRPC.LogResponseData = enabled
	}
}

// WithIgnoreGRPCMethods sets the gRPC method patterns to ignore.
func WithIgnoreGRPCMethods(patterns ...string) Option {
	return func(c *Config) {
		c.GRPC.IgnoreMethods = append(c.GRPC.IgnoreMethods, patterns...)
	}
}

// createInterceptorConfig creates a configuration for gRPC interceptors.
func createInterceptorConfig(config *Config, requestStore store.Store) *internal_grpc.InterceptorConfig {
	return &internal_grpc.InterceptorConfig{
		Store:           requestStore,
		LogRequestData:  config.GRPC.LogRequestData,
		LogResponseData: config.GRPC.LogResponseData,
		IgnoreMethods:   config.GRPC.IgnoreMethods,
	}
}

// WrapGRPCServer wraps a gRPC server with interceptors for request visualization.
func WrapGRPCServer(opts ...Option) []grpc.ServerOption {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// If gRPC support is not enabled, return empty options
	if !config.GRPC.EnableGRPC {
		return []grpc.ServerOption{}
	}

	// Use shared store if provided, otherwise create a new one
	var requestStore store.Store
	var err error

	if config.SharedStore != nil {
		requestStore = config.SharedStore
	} else {
		// Create store based on configuration
		storeConfig := &store.StorageConfig{
			Type:             config.StorageType,
			Capacity:         config.MaxRequests,
			ConnectionString: config.ConnectionString,
			TableName:        config.TableName,
			TTL:              config.RedisTTL,
			ExistingDB:       config.ExistingDB,
		}

		requestStore, err = store.NewStore(storeConfig)
		if err != nil {
			log.Printf("Failed to create configured storage backend for gRPC: %v. Falling back to in-memory storage.", err)
			requestStore = store.NewInMemoryStore(config.MaxRequests)
		}
	}

	// Create interceptor configuration
	interceptorConfig := createInterceptorConfig(config, requestStore)

	// Return server options with interceptors
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(internal_grpc.UnaryServerInterceptor(interceptorConfig)),
		grpc.StreamInterceptor(internal_grpc.StreamServerInterceptor(interceptorConfig)),
	}
}

// NewGRPCServer creates a new gRPC server with visualization interceptors.
func NewGRPCServer(opts ...Option) *grpc.Server {
	// Get interceptor options
	interceptorOpts := WrapGRPCServer(opts...)

	// Create and return server
	return grpc.NewServer(interceptorOpts...)
}

// DialGRPCWithVisualizer creates a client connection with visualization interceptors.
func DialGRPCWithVisualizer(target string, opts ...Option) (*grpc.ClientConn, error) {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// If gRPC support is not enabled, use default dialing
	if !config.GRPC.EnableGRPC {
		return grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use shared store if provided, otherwise create a new one
	var requestStore store.Store
	var err error

	if config.SharedStore != nil {
		requestStore = config.SharedStore
	} else {
		// Create store based on configuration
		storeConfig := &store.StorageConfig{
			Type:             config.StorageType,
			Capacity:         config.MaxRequests,
			ConnectionString: config.ConnectionString,
			TableName:        config.TableName,
			TTL:              config.RedisTTL,
			ExistingDB:       config.ExistingDB,
		}

		requestStore, err = store.NewStore(storeConfig)
		if err != nil {
			log.Printf("Failed to create configured storage backend for gRPC client: %v. Falling back to in-memory storage.", err)
			requestStore = store.NewInMemoryStore(config.MaxRequests)
		}
	}

	// Create interceptor configuration
	interceptorConfig := createInterceptorConfig(config, requestStore)

	// Create client connection with interceptors
	return grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(internal_grpc.UnaryClientInterceptor(interceptorConfig)),
		grpc.WithStreamInterceptor(internal_grpc.StreamClientInterceptor(interceptorConfig)),
	)
}

// GRPCDialContext creates a client connection with context.
func GRPCDialContext(ctx context.Context, target string, opts ...Option) (*grpc.ClientConn, error) {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// If gRPC support is not enabled, use default dialing
	if !config.GRPC.EnableGRPC {
		return grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use shared store if provided, otherwise create a new one
	var requestStore store.Store
	var err error

	if config.SharedStore != nil {
		requestStore = config.SharedStore
	} else {
		// Create store based on configuration
		storeConfig := &store.StorageConfig{
			Type:             config.StorageType,
			Capacity:         config.MaxRequests,
			ConnectionString: config.ConnectionString,
			TableName:        config.TableName,
			TTL:              config.RedisTTL,
			ExistingDB:       config.ExistingDB,
		}

		requestStore, err = store.NewStore(storeConfig)
		if err != nil {
			log.Printf("Failed to create configured storage backend for gRPC client: %v. Falling back to in-memory storage.", err)
			requestStore = store.NewInMemoryStore(config.MaxRequests)
		}
	}

	// Create interceptor configuration
	interceptorConfig := createInterceptorConfig(config, requestStore)

	// Create client connection with interceptors
	return grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(internal_grpc.UnaryClientInterceptor(interceptorConfig)),
		grpc.WithStreamInterceptor(internal_grpc.StreamClientInterceptor(interceptorConfig)),
	)
}
