package agent

import (
	"time"

	"github.com/doganarif/govisual/internal/model"
)

// WithGRPCBatchingEnabled enables or disables request batching for gRPC agent.
// This is a type-safe wrapper around WithBatchingEnabled for gRPC agents.
func WithGRPCBatchingEnabled(enabled bool) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.BatchingEnabled = enabled
	}
}

// WithGRPCBatchSize sets the maximum number of requests in a batch for gRPC agent.
// This is a type-safe wrapper around WithBatchSize for gRPC agents.
func WithGRPCBatchSize(size int) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.BatchSize = size
	}
}

// WithGRPCBatchInterval sets the maximum time to wait before sending a batch for gRPC agent.
// This is a type-safe wrapper around WithBatchInterval for gRPC agents.
func WithGRPCBatchInterval(interval time.Duration) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.BatchInterval = interval
	}
}

// WithGRPCFilter sets a filter function for the gRPC agent.
// This is a type-safe wrapper around WithFilter for gRPC agents.
func WithGRPCFilter(filter func(*model.RequestLog) bool) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.Filter = filter
	}
}

// WithGRPCProcessor sets a processor function for the gRPC agent.
// This is a type-safe wrapper around WithProcessor for gRPC agents.
func WithGRPCProcessor(processor func(*model.RequestLog) *model.RequestLog) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.Processor = processor
	}
}

// WithGRPCMaxBufferSize sets the maximum buffer size for the gRPC agent.
// This is a type-safe wrapper around WithMaxBufferSize for gRPC agents.
func WithGRPCMaxBufferSize(size int) GRPCOption {
	return func(c *GRPCAgentConfig) {
		c.AgentConfig.MaxBufferSize = size
	}
}

// WithHTTPBatchingEnabled enables or disables request batching for HTTP agent.
// This is a type-safe wrapper around WithBatchingEnabled for HTTP agents.
func WithHTTPBatchingEnabled(enabled bool) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.BatchingEnabled = enabled
	}
}

// WithHTTPBatchSize sets the maximum number of requests in a batch for HTTP agent.
// This is a type-safe wrapper around WithBatchSize for HTTP agents.
func WithHTTPBatchSize(size int) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.BatchSize = size
	}
}

// WithHTTPBatchInterval sets the maximum time to wait before sending a batch for HTTP agent.
// This is a type-safe wrapper around WithBatchInterval for HTTP agents.
func WithHTTPBatchInterval(interval time.Duration) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.BatchInterval = interval
	}
}

// WithHTTPFilter sets a filter function for the HTTP agent.
// This is a type-safe wrapper around WithFilter for HTTP agents.
func WithHTTPFilter(filter func(*model.RequestLog) bool) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.Filter = filter
	}
}

// WithHTTPProcessor sets a processor function for the HTTP agent.
// This is a type-safe wrapper around WithProcessor for HTTP agents.
func WithHTTPProcessor(processor func(*model.RequestLog) *model.RequestLog) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.Processor = processor
	}
}

// WithHTTPMaxBufferSize sets the maximum buffer size for the HTTP agent.
// This is a type-safe wrapper around WithMaxBufferSize for HTTP agents.
func WithHTTPMaxBufferSize(size int) HTTPOption {
	return func(c *HTTPAgentConfig) {
		c.AgentConfig.MaxBufferSize = size
	}
}
