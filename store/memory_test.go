package store_test

import (
	"testing"

	"github.com/doganarif/govisual/v2/store"
	"github.com/doganarif/govisual/v2/store/storetest"
)

func TestMemoryStore(t *testing.T) {
	storetest.Run(t, store.NewMemory(10))
}
