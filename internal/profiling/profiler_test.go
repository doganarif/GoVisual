package profiling

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestConcurrentCPUProfiling(t *testing.T) {
	profiler := NewProfiler(100)
	profiler.SetEnabled(true)
	profiler.SetProfileType(ProfileCPU)
	profiler.SetThreshold(1 * time.Millisecond) // Very low threshold for testing

	// Simulate concurrent requests
	var wg sync.WaitGroup
	numRequests := 10

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestID string) {
			defer wg.Done()

			ctx := profiler.StartProfiling(context.Background(), requestID)

			// Simulate some work
			time.Sleep(5 * time.Millisecond)

			metrics := profiler.EndProfiling(ctx)

			// For requests that meet the threshold, we should have CPU profile data
			// Note: Only one request will actually get CPU profiling due to global state
			if metrics != nil {
				t.Logf("Request %s got CPU profile data: %d bytes", requestID, len(metrics.CPUProfile))
			}
		}(fmt.Sprintf("req-%d", i))
	}

	wg.Wait()

	// Verify that no CPU profiling is still active
	if profiler.activeCPUProfile != nil {
		t.Errorf("CPU profiling should not be active after all requests completed")
	}
}

// ballastSink keeps the test allocation visible to the compiler.
var ballastSink []byte

func TestMemoryTotalAllocSurvivesGC(t *testing.T) {
	profiler := NewProfiler(10)
	profiler.SetProfileType(ProfileMemory)
	profiler.SetThreshold(0)

	// Allocate before profiling starts so the baseline heap is high, then
	// free it and force a GC: live heap at EndProfiling drops below the
	// baseline, which used to underflow the uint64 delta.
	ballastSink = make([]byte, 64<<20)
	ctx := profiler.StartProfiling(context.Background(), "req-mem")
	ballastSink = nil
	runtime.GC()

	metrics := profiler.EndProfiling(ctx)
	if metrics == nil {
		t.Fatal("expected metrics, got nil")
	}
	if metrics.MemoryTotalAlloc > 1<<40 {
		t.Fatalf("MemoryTotalAlloc = %d, uint64 underflow", metrics.MemoryTotalAlloc)
	}
	for _, b := range metrics.Bottlenecks {
		if b.Type == "memory" {
			t.Fatalf("unexpected memory bottleneck: %+v", b)
		}
	}
}

func TestEndProfilingRacesWithRecorders(t *testing.T) {
	profiler := NewProfiler(10)
	profiler.SetProfileType(ProfileMemory)
	profiler.SetThreshold(0)

	ctx := profiler.StartProfiling(context.Background(), "req-race")

	// A leaked goroutine that keeps recording while the request finishes.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5000; i++ {
			profiler.RecordSQLQuery(ctx, "SELECT 1", time.Millisecond, 1, nil)
			profiler.RecordHTTPCall(ctx, "GET", "http://example", time.Millisecond, 200, 0)
		}
	}()

	metrics := profiler.EndProfiling(ctx)
	if metrics == nil {
		t.Fatal("expected metrics, got nil")
	}
	for i := 0; i < 20; i++ {
		for _, m := range profiler.GetAllMetrics() {
			if _, err := json.Marshal(m); err != nil {
				t.Fatalf("marshal: %v", err)
			}
		}
	}
	wg.Wait()
}
