// Package transport defines the transport mechanisms for GoVisual.
package transport

import (
	"time"

	"github.com/doganarif/govisual/internal/model"
)

// Transport is an interface that defines how request logs are sent
// from agents to the visualization server.
type Transport interface {
	// Send sends a request log to the visualization server.
	Send(log *model.RequestLog) error

	// SendBatch sends multiple request logs in a single batch.
	SendBatch(logs []*model.RequestLog) error

	// Close closes the transport and performs any cleanup.
	Close() error
}

// TransportType represents the type of transport mechanism.
type TransportType string

const (
	// TransportTypeStore represents a shared store transport.
	TransportTypeStore TransportType = "store"

	// TransportTypeNATS represents a NATS transport.
	TransportTypeNATS TransportType = "nats"

	// TransportTypeHTTP represents an HTTP transport.
	TransportTypeHTTP TransportType = "http"
)

// Config contains common configuration for transport mechanisms.
type Config struct {
	// Type is the type of transport.
	Type TransportType

	// Endpoint is the destination for the transport (e.g., NATS server URL, HTTP endpoint).
	Endpoint string

	// Credentials for authenticating with the transport if needed.
	Credentials map[string]string

	// MaxRetries is the maximum number of retries for failed transmissions.
	MaxRetries int

	// RetryWait is the time to wait between retries.
	RetryWait time.Duration

	// BatchSize is the maximum number of logs to send in a single batch.
	BatchSize int

	// Timeout is the maximum time to wait for a transmission to complete.
	Timeout time.Duration

	// BufferSize is the maximum number of logs to buffer when the transport is unavailable.
	BufferSize int
}

// Option is a function that configures a transport.
type Option func(*Config)

// WithEndpoint sets the endpoint for the transport.
func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

// WithCredentials sets the credentials for the transport.
func WithCredentials(credentials map[string]string) Option {
	return func(c *Config) {
		c.Credentials = credentials
	}
}

// WithMaxRetries sets the maximum number of retries for the transport.
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithRetryWait sets the time to wait between retries.
func WithRetryWait(retryWait time.Duration) Option {
	return func(c *Config) {
		c.RetryWait = retryWait
	}
}

// WithBatchSize sets the maximum number of logs in a batch.
func WithBatchSize(batchSize int) Option {
	return func(c *Config) {
		c.BatchSize = batchSize
	}
}

// WithTimeout sets the timeout for the transport.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithBufferSize sets the buffer size for the transport.
func WithBufferSize(bufferSize int) Option {
	return func(c *Config) {
		c.BufferSize = bufferSize
	}
}

// DefaultConfig returns the default transport configuration.
func DefaultConfig() *Config {
	return &Config{
		Type:       TransportTypeStore,
		MaxRetries: 3,
		RetryWait:  time.Second,
		BatchSize:  10,
		Timeout:    5 * time.Second,
		BufferSize: 100,
	}
}
