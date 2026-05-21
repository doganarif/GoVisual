package store

import (
	"regexp"

	"github.com/doganarif/govisual/internal/model"
)

// Store defines the interface for all storage backends
type Store interface {
	// Add stores a new request log. Returns an error so callers can surface
	// storage failures (otherwise the dashboard silently drops entries).
	Add(log *model.RequestLog) error

	// Get retrieves a specific request log by its ID
	Get(id string) (*model.RequestLog, bool)

	// GetAll returns all stored request logs
	GetAll() []*model.RequestLog

	// Clear clears all stored request logs
	Clear() error

	// GetLatest returns the n most recent request logs
	GetLatest(n int) []*model.RequestLog

	// Close closes any open connections
	Close() error
}

// validTableName matches identifiers safe to inject into SQL via fmt.Sprintf.
// Letters, digits and underscore only — no quoting, dots, or whitespace.
var validTableName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// IsValidTableName reports whether tableName is safe to interpolate into a
// SQL statement. It is consulted by every SQL-backed store before any
// query is constructed; never bypass it.
func IsValidTableName(tableName string) bool {
	return validTableName.MatchString(tableName)
}
