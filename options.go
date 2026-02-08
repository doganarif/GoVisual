package govisual

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/profiling"
	"github.com/doganarif/govisual/internal/store"
)

type Config struct {
	MaxRequests int

	DashboardPath string

	LogRequestBody bool

	LogResponseBody bool

	IgnorePaths []string

	// OpenTelemetry configuration
	EnableOpenTelemetry bool

	ServiceName string

	ServiceVersion string

	OTelEndpoint string

	OTelInsecure bool

	OTelExporter string

	// Storage configuration
	StorageType store.StorageType

	// Connection string for database stores
	ConnectionString string

	// TableName for SQL database stores
	TableName string

	// TTL for Redis store in seconds
	RedisTTL int

	// Existing database connection for SQLite
	ExistingDB *sql.DB

	// Performance Profiling configuration
	EnableProfiling bool

	ProfileType profiling.ProfileType

	ProfileThreshold time.Duration

	MaxProfileMetrics int
}

// Option is a function that modifies the configuration
type Option func(*Config)

// WithMaxRequests sets the maximum number of requests to store
func WithMaxRequests(max int) Option {
	return func(c *Config) {
		c.MaxRequests = max
	}
}

// WithDashboardPath sets the path to access the dashboard
func WithDashboardPath(path string) Option {
	return func(c *Config) {
		c.DashboardPath = path
	}
}

// WithRequestBodyLogging enables or disables request body logging
func WithRequestBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogRequestBody = enabled
	}
}

// WithResponseBodyLogging enables or disables response body logging
func WithResponseBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogResponseBody = enabled
	}
}

// WithIgnorePaths sets the path patterns to ignore
func WithIgnorePaths(patterns ...string) Option {
	return func(c *Config) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// WithOpenTelemetry enables or disables OpenTelemetry instrumentation
func WithOpenTelemetry(enabled bool) Option {
	return func(c *Config) {
		c.EnableOpenTelemetry = enabled
	}
}

// WithServiceName sets the service name for OpenTelemetry
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion sets the service version for OpenTelemetry
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithOTelEndpoint sets the OTLP endpoint for exporting telemetry data
func WithOTelEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.OTelEndpoint = endpoint
	}
}

// WithOTelInsecure sets whether to use an insecure connection for OTLP
func WithOTelInsecure(insecure bool) Option {
	return func(c *Config) {
		c.OTelInsecure = insecure
	}
}

// WithOTelExporter sets the type of exporter to use.
// Valid values: "otlp" (default), "stdout" (for debugging), "noop" (for benchmarking)
func WithOTelExporter(exporterType string) Option {
	return func(c *Config) {
		c.OTelExporter = exporterType
	}
}

// WithMemoryStorage configures the application to use in-memory storage
func WithMemoryStorage() Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeMemory
	}
}

// WithPostgresStorage configures the application to use PostgreSQL storage
func WithPostgresStorage(connStr string, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypePostgres
		c.ConnectionString = connStr
		c.TableName = tableName
	}
}

// WithSQLiteStorage configures the application to use SQLite storage
func WithSQLiteStorage(dbPath string, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeSQLite
		c.ConnectionString = dbPath
		c.TableName = tableName
	}
}

// WithSQLiteStorageDB configures the application to use SQLite storage with an existing database connection
func WithSQLiteStorageDB(db *sql.DB, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeSQLiteWithDB
		c.ExistingDB = db
		c.TableName = tableName
	}
}

// WithRedisStorage configures the application to use Redis storage
func WithRedisStorage(connStr string, ttlSeconds int) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeRedis
		c.ConnectionString = connStr
		c.RedisTTL = ttlSeconds
	}
}

// WithMongoDBStorage configures the application to use MongoDB storage
func WithMongoDBStorage(uri, databaseName, collectionName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeMongoDB
		c.ConnectionString = uri
		c.TableName = fmt.Sprintf("%s.%s", databaseName, collectionName)
	}
}

// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
func (c *Config) ShouldIgnorePath(path string) bool {
	// First check if it's the dashboard path which should always be ignored to prevent recursive logging
	if path == c.DashboardPath || strings.HasPrefix(path, c.DashboardPath+"/") {
		return true
	}

	// Then check against provided ignore patterns
	for _, pattern := range c.IgnorePaths {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Special handling for path groups with trailing slash
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			// If pattern ends with /, check if path starts with pattern
			if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
				return true
			}
		}
	}

	return false
}

// WithProfiling enables or disables performance profiling
func WithProfiling(enabled bool) Option {
	return func(c *Config) {
		c.EnableProfiling = enabled
	}
}

// WithProfileType sets the types of profiling to perform
func WithProfileType(profileType profiling.ProfileType) Option {
	return func(c *Config) {
		c.ProfileType = profileType
	}
}

// WithProfileThreshold sets the minimum duration to trigger profiling
func WithProfileThreshold(threshold time.Duration) Option {
	return func(c *Config) {
		c.ProfileThreshold = threshold
	}
}

// WithMaxProfileMetrics sets the maximum number of profile metrics to store
func WithMaxProfileMetrics(max int) Option {
	return func(c *Config) {
		c.MaxProfileMetrics = max
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		MaxRequests:         100,
		DashboardPath:       "/__viz",
		LogRequestBody:      false,
		LogResponseBody:     false,
		IgnorePaths:         []string{},
		EnableOpenTelemetry: false,
		ServiceName:         "govisual",
		ServiceVersion:      "dev",
		OTelEndpoint:        "localhost:4317",
		OTelInsecure:        true,
		OTelExporter:        "otlp",
		StorageType:         store.StorageTypeMemory,
		TableName:           "govisual_requests",
		RedisTTL:            86400, // 24 hours
		EnableProfiling:     false,
		ProfileType:         profiling.ProfileAll,
		ProfileThreshold:    10 * time.Millisecond,
		MaxProfileMetrics:   1000,
	}
}
