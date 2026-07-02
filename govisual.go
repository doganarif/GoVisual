// Package govisual wraps an http.Handler with request capture and serves a
// debugging dashboard for local development.
package govisual

import (
	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

// Store is the storage interface for captured requests. See the store
// package for the in-memory implementation and store/* modules for
// database-backed ones.
type Store = store.Store

// RequestLog is a single captured request.
type RequestLog = store.RequestLog

// ProfileType selects which profile kinds the profiler collects.
type ProfileType = profiling.ProfileType

const (
	ProfileCPU       = profiling.ProfileCPU
	ProfileMemory    = profiling.ProfileMemory
	ProfileGoroutine = profiling.ProfileGoroutine
	ProfileAll       = profiling.ProfileAll
)
