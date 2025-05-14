package store

import "github.com/doganarif/govisual/internal/model"

// Store defines the interface for all storage backends
type Store interface {
	// Add adds a new request log to the store
	Add(log *model.RequestLog)

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
