package agent

import (
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/transport"
)

// Option is a function that configures an agent.
type Option func(*AgentConfig)

// Apply applies the option to an AgentConfig
func (o Option) Apply(config *AgentConfig) {
	o(config)
}

// ForGRPC converts a base option to a GRPC option
func (o Option) ForGRPC() GRPCOption {
	return func(c *GRPCAgentConfig) {
		o(&c.AgentConfig)
	}
}

// ForHTTP converts a base option to an HTTP option
func (o Option) ForHTTP() HTTPOption {
	return func(c *HTTPAgentConfig) {
		o(&c.AgentConfig)
	}
}

// WithTransport sets the transport mechanism for the agent.
func WithTransport(transport transport.Transport) Option {
	return func(c *AgentConfig) {
		c.Transport = transport
	}
}

// WithMaxBufferSize sets the maximum number of requests to buffer.
func WithMaxBufferSize(size int) Option {
	return func(c *AgentConfig) {
		c.MaxBufferSize = size
	}
}

// WithBatchingEnabled enables or disables request batching.
func WithBatchingEnabled(enabled bool) Option {
	return func(c *AgentConfig) {
		c.BatchingEnabled = enabled
	}
}

// WithBatchSize sets the maximum number of requests in a batch.
func WithBatchSize(size int) Option {
	return func(c *AgentConfig) {
		c.BatchSize = size
	}
}

// WithBatchInterval sets the maximum time to wait before sending a batch.
func WithBatchInterval(interval time.Duration) Option {
	return func(c *AgentConfig) {
		c.BatchInterval = interval
	}
}

// WithFilter sets a filter function for the agent.
func WithFilter(filter func(*model.RequestLog) bool) Option {
	return func(c *AgentConfig) {
		c.Filter = filter
	}
}

// WithProcessor sets a processor function for the agent.
func WithProcessor(processor func(*model.RequestLog) *model.RequestLog) Option {
	return func(c *AgentConfig) {
		c.Processor = processor
	}
}
