package profiling

import (
	"context"
	"fmt"
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
