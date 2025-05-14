package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/doganarif/govisual/internal/model"
)

// HTTPTransport is a transport that sends request logs via HTTP.
type HTTPTransport struct {
	config       *Config
	client       *http.Client
	buffer       []*model.RequestLog
	bufferMutex  sync.Mutex
	shutdown     chan struct{}
	wg           sync.WaitGroup
	batchProcess bool
}

// NewHTTPTransport creates a new HTTP transport.
func NewHTTPTransport(endpoint string, opts ...Option) *HTTPTransport {
	config := DefaultConfig()
	config.Type = TransportTypeHTTP
	config.Endpoint = endpoint

	for _, opt := range opts {
		opt(config)
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	t := &HTTPTransport{
		config:       config,
		client:       client,
		buffer:       make([]*model.RequestLog, 0, config.BufferSize),
		shutdown:     make(chan struct{}),
		batchProcess: config.BatchSize > 1,
	}

	// Start background processor if batching is enabled
	if t.batchProcess {
		t.startBackgroundProcessor()
	}

	return t
}

// Send sends a request log via HTTP.
func (t *HTTPTransport) Send(log *model.RequestLog) error {
	// If batching is enabled, add to buffer
	if t.batchProcess {
		t.bufferMutex.Lock()
		defer t.bufferMutex.Unlock()

		// If buffer is full, try to send immediately
		if len(t.buffer) >= t.config.BufferSize {
			logs := make([]*model.RequestLog, len(t.buffer))
			copy(logs, t.buffer)
			t.buffer = t.buffer[:0]
			go t.sendBatchWithRetry(logs)
		}

		// Add log to buffer
		t.buffer = append(t.buffer, log)
		return nil
	}

	// Send single log immediately
	return t.sendSingleWithRetry(log)
}

// SendBatch sends multiple request logs in a single batch.
func (t *HTTPTransport) SendBatch(logs []*model.RequestLog) error {
	if len(logs) == 0 {
		return nil
	}

	return t.sendBatchWithRetry(logs)
}

// Close closes the HTTP transport.
func (t *HTTPTransport) Close() error {
	close(t.shutdown)
	t.wg.Wait()

	// Send any remaining logs
	t.bufferMutex.Lock()
	defer t.bufferMutex.Unlock()

	if len(t.buffer) > 0 {
		return t.sendBatchWithRetry(t.buffer)
	}

	return nil
}

// startBackgroundProcessor starts a goroutine that periodically processes the buffer.
func (t *HTTPTransport) startBackgroundProcessor() {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.processBatch()
			case <-t.shutdown:
				return
			}
		}
	}()
}

// processBatch processes the buffer.
func (t *HTTPTransport) processBatch() {
	t.bufferMutex.Lock()
	defer t.bufferMutex.Unlock()

	if len(t.buffer) == 0 {
		return
	}

	// Create a copy of the buffer
	logs := make([]*model.RequestLog, len(t.buffer))
	copy(logs, t.buffer)
	t.buffer = t.buffer[:0]

	// Send batch
	go t.sendBatchWithRetry(logs)
}

// sendSingleWithRetry sends a single log with retries.
func (t *HTTPTransport) sendSingleWithRetry(log *model.RequestLog) error {
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			time.Sleep(t.config.RetryWait)
		}

		req, err := http.NewRequest(http.MethodPost, t.config.Endpoint, bytes.NewReader(data))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		t.addAuthHeaders(req)

		// Set a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), t.config.Timeout)
		req = req.WithContext(ctx)

		resp, err := t.client.Do(req)
		cancel()

		if err != nil {
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
	}

	return fmt.Errorf("failed to send log after %d attempts", t.config.MaxRetries+1)
}

// sendBatchWithRetry sends multiple logs with retries.
func (t *HTTPTransport) sendBatchWithRetry(logs []*model.RequestLog) error {
	data, err := json.Marshal(logs)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			time.Sleep(t.config.RetryWait)
		}

		req, err := http.NewRequest(http.MethodPost, t.config.Endpoint+"/batch", bytes.NewReader(data))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		t.addAuthHeaders(req)

		// Set a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), t.config.Timeout)
		req = req.WithContext(ctx)

		resp, err := t.client.Do(req)
		cancel()

		if err != nil {
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
	}

	return fmt.Errorf("failed to send batch after %d attempts", t.config.MaxRetries+1)
}

// addAuthHeaders adds authentication headers to the request.
func (t *HTTPTransport) addAuthHeaders(req *http.Request) {
	if t.config.Credentials != nil {
		if token, ok := t.config.Credentials["token"]; ok {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		if apiKey, ok := t.config.Credentials["api_key"]; ok {
			req.Header.Set("X-API-Key", apiKey)
		}
	}
}
