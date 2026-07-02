// Package store defines the request log model and the storage interface
// govisual persists captured requests through. The in-memory store lives
// here; database-backed stores are separate modules under store/ so their
// drivers stay out of builds that don't use them.
package store

import "regexp"

// Store is the interface all storage backends implement.
type Store interface {
	// Add stores a new request log. Returns an error so callers can surface
	// storage failures (otherwise the dashboard silently drops entries).
	Add(log *RequestLog) error

	// Get retrieves a specific request log by its ID
	Get(id string) (*RequestLog, bool)

	// GetAll returns all stored request logs
	GetAll() []*RequestLog

	// Clear clears all stored request logs
	Clear() error

	// GetLatest returns the n most recent request logs
	GetLatest(n int) []*RequestLog

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
