package govisual

import (
	"database/sql"
	"fmt"

	"github.com/doganarif/govisual/internal/options"
	"github.com/doganarif/govisual/internal/store"
)

// Option is a function that modifies the configuration
type Option func(*options.Config)

// WithMaxRequests sets the maximum number of requests to store
func WithMaxRequests(max int) Option {
	return func(c *options.Config) {
		c.MaxRequests = max
	}
}

// WithDashboardPath sets the path to access the dashboard
func WithDashboardPath(path string) Option {
	return func(c *options.Config) {
		c.DashboardPath = path
	}
}

// WithRequestBodyLogging enables or disables request body logging
func WithRequestBodyLogging(enabled bool) Option {
	return func(c *options.Config) {
		c.LogRequestBody = enabled
	}
}

// WithResponseBodyLogging enables or disables response body logging
func WithResponseBodyLogging(enabled bool) Option {
	return func(c *options.Config) {
		c.LogResponseBody = enabled
	}
}

func WithConsoleLogging(enabled bool) Option {
	return func(c *options.Config) {
		c.LogRequestToConsole = enabled
	}
}

// WithIgnorePaths sets the path patterns to ignore
func WithIgnorePaths(patterns ...string) Option {
	return func(c *options.Config) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// WithOpenTelemetry enables or disables OpenTelemetry instrumentation
func WithOpenTelemetry(enabled bool) Option {
	return func(c *options.Config) {
		c.EnableOpenTelemetry = enabled
	}
}

// WithServiceName sets the service name for OpenTelemetry
func WithServiceName(name string) Option {
	return func(c *options.Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion sets the service version for OpenTelemetry
func WithServiceVersion(version string) Option {
	return func(c *options.Config) {
		c.ServiceVersion = version
	}
}

// WithOTelEndpoint sets the OTLP endpoint for exporting telemetry data
func WithOTelEndpoint(endpoint string) Option {
	return func(c *options.Config) {
		c.OTelEndpoint = endpoint
	}
}

// WithMemoryStorage configures the application to use in-memory storage
func WithMemoryStorage() Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypeMemory
	}
}

// WithPostgresStorage configures the application to use PostgreSQL storage
func WithPostgresStorage(connStr string, tableName string) Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypePostgres
		c.ConnectionString = connStr
		c.TableName = tableName
	}
}

// WithSQLiteStorage configures the application to use SQLite storage
func WithSQLiteStorage(dbPath string, tableName string) Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypeSQLite
		c.ConnectionString = dbPath
		c.TableName = tableName
	}
}

// WithSQLiteStorageDB configures the application to use SQLite storage with an existing database connection
func WithSQLiteStorageDB(db *sql.DB, tableName string) Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypeSQLiteWithDB
		c.ExistingDB = db
		c.TableName = tableName
	}
}

// WithRedisStorage configures the application to use Redis storage
func WithRedisStorage(connStr string, ttlSeconds int) Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypeRedis
		c.ConnectionString = connStr
		c.RedisTTL = ttlSeconds
	}
}

// WithMongoDBStorage configures the application to use MongoDB storage
func WithMongoDBStorage(uri, databaseName, collectionName string) Option {
	return func(c *options.Config) {
		c.StorageType = store.StorageTypeMongoDB
		c.ConnectionString = uri
		c.TableName = fmt.Sprintf("%s.%s", databaseName, collectionName)
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *options.Config {
	return &options.Config{
		MaxRequests:         100,
		DashboardPath:       "/__viz",
		LogRequestBody:      false,
		LogResponseBody:     false,
		LogRequestToConsole: false,
		IgnorePaths:         []string{},
		EnableOpenTelemetry: false,
		ServiceName:         "govisual",
		ServiceVersion:      "dev",
		OTelEndpoint:        "localhost:4317",
		StorageType:         store.StorageTypeMemory,
		TableName:           "govisual_requests",
		RedisTTL:            86400, // 24 hours
	}
}
