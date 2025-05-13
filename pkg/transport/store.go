// Package transport defines the transport mechanisms for GoVisual.
package transport

import (
	"github.com/doganarif/govisual/internal/model"
	"github.com/doganarif/govisual/internal/store"
)

// StoreTransport is a transport that uses a shared store for communication.
type StoreTransport struct {
	store store.Store
}

// NewStoreTransport creates a new store transport with the given store.
func NewStoreTransport(store store.Store) *StoreTransport {
	return &StoreTransport{
		store: store,
	}
}

// Send sends a request log to the store.
func (t *StoreTransport) Send(log *model.RequestLog) error {
	t.store.Add(log)
	return nil
}

// SendBatch sends multiple request logs to the store.
func (t *StoreTransport) SendBatch(logs []*model.RequestLog) error {
	for _, log := range logs {
		t.store.Add(log)
	}
	return nil
}

// Close closes the store transport.
func (t *StoreTransport) Close() error {
	if t.store != nil {
		return t.store.Close()
	}
	return nil
}
