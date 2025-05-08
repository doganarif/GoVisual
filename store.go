package govisual

import (
	"github.com/doganarif/govisual/internal/store"
)

// Store is a public interface for the storage backend.
// It exposes only the methods that should be available to the client code.
type Store interface {
	// Close closes any open connections.
	Close() error
}

// NewStore creates a shared store that can be used with both HTTP and gRPC visualizers.
// This allows gRPC requests to be visible in the same dashboard as HTTP requests.
func NewStore(opts ...Option) (Store, error) {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create store based on configuration
	storeConfig := &store.StorageConfig{
		Type:             config.StorageType,
		Capacity:         config.MaxRequests,
		ConnectionString: config.ConnectionString,
		TableName:        config.TableName,
		TTL:              config.RedisTTL,
		ExistingDB:       config.ExistingDB,
	}

	return store.NewStore(storeConfig)
}

// WithSharedStore configures the middleware to use a shared store.
func WithSharedStore(sharedStore Store) Option {
	return func(c *Config) {
		if internalStore, ok := sharedStore.(store.Store); ok {
			c.SharedStore = internalStore
		}
	}
}
