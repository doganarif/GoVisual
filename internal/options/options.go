package options

import (
	"database/sql"
	"path/filepath"
	"strings"

	"github.com/doganarif/govisual/internal/store"
)

type Config struct {
	MaxRequests int

	DashboardPath string

	LogRequestBody bool

	LogResponseBody bool

	LogRequestToConsole bool

	IgnorePaths []string

	// OpenTelemetry configuration
	EnableOpenTelemetry bool

	ServiceName string

	ServiceVersion string

	OTelEndpoint string

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
}

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
