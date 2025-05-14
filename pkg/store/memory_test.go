package store

import (
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore(10)
	runStoreTests(t, store)
}
