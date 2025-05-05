package store

import (
	"fmt"
)

// StorageType represents the type of storage backend to use
type StorageType string

const (
	// StorageTypeMemory represents in-memory storage
	StorageTypeMemory StorageType = "memory"

	// StorageTypePostgres represents PostgreSQL storage
	StorageTypePostgres StorageType = "postgres"

	// StorageTypeRedis represents Redis storage
	StorageTypeRedis StorageType = "redis"

	// StorageTypeSQLite represents SQLite storage
	StorageTypeSQLite StorageType = "sqlite"
)

// StorageConfig represents configuration options for storage backends
type StorageConfig struct {
	// Type specifies which storage backend to use
	Type StorageType

	// Capacity is the maximum number of requests to store (applicable to memory store)
	Capacity int

	// ConnectionString is the database connection string (applicable to DB stores)
	ConnectionString string

	// TableName is the table name for SQL databases
	TableName string

	// TTL is the time-to-live for entries in Redis
	TTL int
}

// NewStore creates a new storage backend based on configuration
func NewStore(config *StorageConfig) (Store, error) {
	switch config.Type {
	case StorageTypeMemory:
		return NewInMemoryStore(config.Capacity), nil

	case StorageTypePostgres:
		return NewPostgresStore(config.ConnectionString, config.TableName, config.Capacity)

	case StorageTypeRedis:
		return NewRedisStore(config.ConnectionString, config.Capacity, config.TTL)

	case StorageTypeSQLite:
		return NewSQLiteStore(config.ConnectionString, config.TableName, config.Capacity)

	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}
