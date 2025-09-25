# Performance Profiling Example

This example demonstrates GoVisual's performance profiling capabilities, including CPU profiling, memory tracking, bottleneck detection, and flame graph visualization.

## Features Demonstrated

- **CPU Profiling**: Track CPU-intensive operations
- **Memory Profiling**: Monitor memory allocations and garbage collection
- **SQL Query Tracking**: Record database query performance
- **HTTP Call Monitoring**: Track external API calls
- **Bottleneck Detection**: Automatically identify performance issues
- **Flame Graph Visualization**: Interactive visualization of CPU profiles

## Running the Example

```bash
cd cmd/examples/profiling
go run main.go
```

Then open your browser to:

- Application: http://localhost:8080
- GoVisual Dashboard: http://localhost:8080/\_\_viz

## Available Endpoints

The example provides several endpoints with different performance characteristics:

### 1. Fast Endpoint (`/api/fast`)

- **Response Time**: < 10ms
- **Characteristics**: Minimal processing, quick JSON response
- **Use Case**: Baseline for comparison

### 2. Slow Endpoint (`/api/slow`)

- **Response Time**: 100-500ms
- **Characteristics**: Simulated delay
- **Use Case**: Testing slow request handling

### 3. CPU Intensive (`/api/cpu-intensive`)

- **Response Time**: Variable
- **Characteristics**: Prime number calculation using Sieve of Eratosthenes
- **Use Case**: Testing CPU profiling and flame graphs

### 4. Memory Intensive (`/api/memory-intensive`)

- **Response Time**: Variable
- **Characteristics**: Large memory allocations
- **Use Case**: Testing memory profiling and GC impact

### 5. Database Operations (`/api/database`)

- **Response Time**: Variable
- **Characteristics**: Multiple SQL queries
- **Use Case**: Testing SQL query tracking

### 6. External API (`/api/external`)

- **Response Time**: Variable (depends on network)
- **Characteristics**: Makes real external HTTP calls
- **Use Case**: Testing HTTP call monitoring

### 7. Complex Workload (`/api/complex`)

- **Response Time**: Variable
- **Characteristics**: Combines CPU, memory, database, and external calls
- **Use Case**: Testing bottleneck detection with mixed workloads

## Using the Performance Features

### 1. Generate Test Traffic

Click on the various endpoint buttons on the home page or use curl:

```bash
# Fast endpoint
curl http://localhost:8080/api/fast

# CPU intensive
curl http://localhost:8080/api/cpu-intensive

# Database operations
curl http://localhost:8080/api/database

# Complex workload
curl http://localhost:8080/api/complex
```

### 2. View Performance Metrics

1. Open the GoVisual dashboard at http://localhost:8080/\_\_viz
2. Click on any request in the request table
3. Click the "View Performance" button (green button in request details)
4. Explore the performance tabs:
   - **Bottlenecks**: Identified performance issues
   - **Flame Graph**: CPU profile visualization
   - **SQL Queries**: Database query performance
   - **HTTP Calls**: External API call metrics
   - **Function Timings**: Individual function durations

### 3. Understanding Bottlenecks

The system automatically detects bottlenecks based on:

- **Database**: When SQL queries take > 50% of request time
- **HTTP**: When external calls take > 33% of request time
- **Memory**: When > 10MB is allocated
- **GC**: When garbage collection takes > 10% of request time
- **CPU**: When specific functions take > 25% of request time

### 4. Interpreting Flame Graphs

The flame graph shows:

- **Width**: Represents time spent in function
- **Height**: Represents call stack depth
- **Colors**: Different functions (consistent coloring)
- **Hover**: Shows function details and percentage

### 5. Performance Metrics

Key metrics displayed:

- **CPU Time**: Total CPU time consumed
- **Memory Allocated**: Total memory allocated during request
- **Goroutines**: Number of active goroutines
- **GC Pauses**: Time spent in garbage collection

## Configuration Options

The example uses these profiling configurations:

```go
govisual.WithProfiling(true),                      // Enable profiling
govisual.WithProfileType(profiling.ProfileAll),    // Profile everything
govisual.WithProfileThreshold(5*time.Millisecond), // Only profile requests > 5ms
govisual.WithMaxProfileMetrics(500),               // Store up to 500 metrics
```

### Profile Types

You can configure what to profile:

```go
// Profile only CPU
govisual.WithProfileType(profiling.ProfileCPU)

// Profile CPU and Memory
govisual.WithProfileType(profiling.ProfileCPU | profiling.ProfileMemory)

// Profile everything
govisual.WithProfileType(profiling.ProfileAll)
```

### Performance Threshold

Set minimum duration to trigger profiling:

```go
// Only profile requests taking > 100ms
govisual.WithProfileThreshold(100 * time.Millisecond)

// Profile all requests
govisual.WithProfileThreshold(0)
```

## Best Practices

1. **Development vs Production**:

   - Enable full profiling in development
   - Use sampling or higher thresholds in production
   - Consider the overhead of profiling

2. **Threshold Configuration**:

   - Set appropriate thresholds based on your SLAs
   - Start with higher thresholds and adjust down as needed
   - Monitor the overhead of profiling itself

3. **Storage Limits**:

   - Configure `MaxProfileMetrics` based on memory constraints
   - Older metrics are automatically evicted when limit is reached
   - Consider external storage for long-term analysis

4. **Bottleneck Analysis**:
   - Focus on bottlenecks with highest impact percentages
   - Address database and external API bottlenecks first
   - Use caching to reduce repeated expensive operations

## Performance Impact

The profiling system is designed for minimal overhead:

- **CPU Profiling**: ~5-10% overhead when enabled
- **Memory Profiling**: < 5% overhead
- **Disabled State**: Near-zero overhead (atomic boolean check)
- **Storage**: O(n) where n = MaxProfileMetrics

## Troubleshooting

### No Performance Metrics Showing

1. Ensure profiling is enabled in configuration
2. Check that request duration exceeds threshold
3. Verify the request completed successfully

### Flame Graph Not Displaying

1. CPU profiling must be enabled
2. Request must have sufficient CPU usage
3. Check browser console for errors

### High Memory Usage

1. Reduce `MaxProfileMetrics` setting
2. Increase profile threshold
3. Disable memory profiling if not needed

## Advanced Usage

### Custom Profiling in Your Code

You can integrate with the profiler in your handlers:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get profiler from context (if available)
    if prof, ok := ctx.Value("profiler").(*profiling.Profiler); ok {
        // Record custom SQL query
        prof.RecordSQLQuery(ctx, query, duration, rows, err)

        // Record custom HTTP call
        prof.RecordHTTPCall(ctx, method, url, duration, status, size)

        // Record custom function timing
        prof.RecordFunction(ctx, "myExpensiveOperation", func() error {
            // Your code here
            return nil
        })
    }
}
```

## Next Steps

- Try different combinations of profile types
- Experiment with threshold settings
- Create custom bottleneck scenarios
- Integrate profiling into your own applications
