// Package agent defines the core agent functionality for GoVisual.
package agent

import (
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/pkg/transport"
)

// Agent represents a GoVisual data collection agent that can be
// attached to various service types (gRPC, HTTP, etc.) to collect
// request/response information.
type Agent interface {
	// Process takes a request log and sends it through the configured
	// transport mechanism.
	Process(log *model.RequestLog) error

	// Close shuts down the agent and performs any cleanup operations.
	Close() error

	// Type returns the agent type (e.g., "grpc", "http").
	Type() string
}

// AgentConfig contains the common configuration options for all agent types.
type AgentConfig struct {
	// Transport is the mechanism used to send data to the visualization server.
	Transport transport.Transport

	// MaxBufferSize is the maximum number of requests to buffer if the
	// transport is unavailable.
	MaxBufferSize int

	// BatchingEnabled determines whether requests should be batched
	// before being sent to the transport.
	BatchingEnabled bool

	// BatchSize is the maximum number of requests to send in a single batch.
	BatchSize int

	// BatchInterval is the maximum time to wait before sending a batch.
	BatchInterval time.Duration

	// Filter is a function that determines whether a request should be processed.
	Filter func(*model.RequestLog) bool

	// Processor is a function that modifies the request log before transport.
	Processor func(*model.RequestLog) *model.RequestLog
}

// BaseAgent implements the common functionality for all agent types.
type BaseAgent struct {
	config    AgentConfig
	agentType string
}

// NewBaseAgent creates a new agent with the given configuration.
func NewBaseAgent(agentType string, config AgentConfig) *BaseAgent {
	// Set defaults if not provided
	if config.MaxBufferSize <= 0 {
		config.MaxBufferSize = 100
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	if config.BatchInterval <= 0 {
		config.BatchInterval = 5 * time.Second
	}

	return &BaseAgent{
		config:    config,
		agentType: agentType,
	}
}

// Process processes a request log and sends it through the transport.
func (a *BaseAgent) Process(log *model.RequestLog) error {
	// Apply filter if exists
	if a.config.Filter != nil && !a.config.Filter(log) {
		return nil
	}

	// Apply processor if exists
	if a.config.Processor != nil {
		log = a.config.Processor(log)
	}

	// Send through transport
	return a.config.Transport.Send(log)
}

// Close closes the agent and its transport.
func (a *BaseAgent) Close() error {
	if a.config.Transport != nil {
		return a.config.Transport.Close()
	}
	return nil
}

// Type returns the agent type.
func (a *BaseAgent) Type() string {
	return a.agentType
}
