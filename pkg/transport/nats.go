package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/doganarif/govisual/internal"
	"github.com/doganarif/govisual/internal/model"
	"github.com/nats-io/nats.go"
)

// NATSTransport is a transport that uses NATS for communication.
type NATSTransport struct {
	ctx    context.Context
	cancel context.CancelFunc

	config      *Config
	conn        *nats.Conn
	buffer      []*model.RequestLog
	bufferMutex sync.Mutex
	wg          sync.WaitGroup
}

// NewNATSTransport creates a new NATS transport with the given options.
func NewNATSTransport(serverURL string, opts ...Option) (*NATSTransport, error) {
	config := DefaultConfig()
	config.Type = TransportTypeNATS
	config.Endpoint = serverURL

	for _, opt := range opts {
		opt(config)
	}

	// Connect to NATS
	natsOpts := []nats.Option{
		nats.Name("GoVisual Agent"),
		nats.Timeout(config.Timeout),
		nats.ReconnectWait(config.RetryWait),
		nats.MaxReconnects(config.MaxRetries),
	}

	// Add credentials if provided
	if config.Credentials != nil {
		if username, ok := config.Credentials["username"]; ok {
			if password, ok := config.Credentials["password"]; ok {
				natsOpts = append(natsOpts, nats.UserInfo(username, password))
			}
		}

		if token, ok := config.Credentials["token"]; ok {
			natsOpts = append(natsOpts, nats.Token(token))
		}
	}

	conn, err := nats.Connect(serverURL, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}
	if !conn.IsConnected() {
		return nil, fmt.Errorf("NATS did not connect")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t := &NATSTransport{
		ctx:    ctx,
		cancel: cancel,
		config: config,
		conn:   conn,
		buffer: make([]*model.RequestLog, 0, config.BufferSize),
	}

	// Start the background processor for batching
	if config.BatchSize > 1 {
		t.startBackgroundProcessor()
	}

	return t, nil
}

// Send sends a request log to NATS.
func (t *NATSTransport) Send(log *model.RequestLog) error {
	// If batching is enabled, add to buffer
	if t.config.BatchSize > 1 {
		t.bufferMutex.Lock()
		defer t.bufferMutex.Unlock()

		// If buffer is full, try to send immediately
		if len(t.buffer) >= t.config.BufferSize {
			logs := make([]*model.RequestLog, len(t.buffer))
			copy(logs, t.buffer)
			t.buffer = t.buffer[:0]
			go t.SendBatch(logs)
		}

		// Add log to buffer
		t.buffer = append(t.buffer, log)
		return nil
	}

	// Send single log immediately
	return t.sendSingle(log)
}

// SendBatch sends multiple request logs in a single batch.
func (t *NATSTransport) SendBatch(logs []*model.RequestLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Serialize the batch
	data, err := json.Marshal(logs)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	// Send to NATS
	return t.conn.Publish(internal.NatsSubjectBatchLogMessages, data)
}

// Close closes the NATS transport.
func (t *NATSTransport) Close() error {
	t.cancel()
	t.wg.Wait()

	// Send any remaining logs
	t.bufferMutex.Lock()
	logs := t.buffer
	t.buffer = nil
	t.bufferMutex.Unlock()

	if len(logs) > 0 {
		if err := t.SendBatch(logs); err != nil {
			t.conn.Close()
			return err
		}
	}

	t.conn.Close()
	return nil
}

// startBackgroundProcessor starts a goroutine that periodically processes the buffer.
func (t *NATSTransport) startBackgroundProcessor() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.processBatch()
			case <-t.ctx.Done():
				return
			}
		}
	}()
}

// processBatch processes the buffer if it contains enough logs or if enough time has passed.
func (t *NATSTransport) processBatch() {
	t.bufferMutex.Lock()
	defer t.bufferMutex.Unlock()

	if len(t.buffer) == 0 {
		return
	}

	// If buffer has enough logs or it's been long enough, send them
	if len(t.buffer) >= t.config.BatchSize {
		logs := make([]*model.RequestLog, len(t.buffer))
		copy(logs, t.buffer)
		t.buffer = t.buffer[:0]
		go t.SendBatch(logs)
	}
}

// sendSingle sends a single request log.
func (t *NATSTransport) sendSingle(log *model.RequestLog) error {
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Send to NATS
	return t.conn.Publish(internal.NatsSubjectSingleLogMessages, data)
}
