package store

import (
	"database/sql"
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

	// StorageTypeSQLiteWithDB represents SQLite storage with existing connection
	StorageTypeSQLiteWithDB StorageType = "sqlite_with_db"
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

	// ExistingDB is an existing database connection (applicable to StorageTypeSQLiteWithDB)
	ExistingDB *sql.DB
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
		// The original SQLite store with automatic driver registration
		return NewSQLiteStore(config.ConnectionString, config.TableName, config.Capacity)

	case StorageTypeSQLiteWithDB:
		// New SQLite store that accepts an existing DB connection
		if config.ExistingDB == nil {
			return nil, fmt.Errorf("existing DB connection is required for sqlite_with_db storage type")
		}
		return NewSQLiteStoreWithDB(config.ExistingDB, config.TableName, config.Capacity)

	default:
		return nil, fmt.Errorf("unknown storage type: %s", config.Type)
	}
}
