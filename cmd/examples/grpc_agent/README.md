# GoVisual Agent Architecture

This document explains how to use the new agent architecture in GoVisual to monitor distributed services.

## What is the Agent Architecture?

The agent architecture allows GoVisual to collect request data from multiple services (potentially running on different machines) and visualize them in a central dashboard. This is particularly useful for microservice architectures or distributed systems where multiple services need to be monitored.

### Key Components

1. **Agents**: Lightweight components that attach to services (gRPC, HTTP) to collect request/response data
2. **Transports**: Mechanisms for sending data from agents to the visualization server
3. **Visualization Server**: Central server that receives, stores, and displays the request data

## Getting Started

### 1. Setting Up the Visualization Server

Start by initializing a visualization server that will display the dashboard and receive agent data:

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/doganarif/govisual"
    "github.com/doganarif/govisual/internal/server"
)

func main() {
    // Create a store for visualization data
    store, err := govisual.NewStore(
        govisual.WithMaxRequests(1000),
        govisual.WithMemoryStorage(),
    )
    if err != nil {
        log.Fatalf("Failed to create store: %v", err)
    }
    
    // Create a server mux
    mux := http.NewServeMux()
    
    // Add homepage
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("<h1>GoVisual Dashboard</h1><p>Go to <a href='/__viz'>/__viz</a> to see the dashboard</p>"))
    })
    
    // Register agent API endpoints
    agentAPI := server.NewAgentAPI(store)
    agentAPI.RegisterHandlers(mux)
    
    // Wrap with GoVisual
    handler := govisual.Wrap(
        mux,
        govisual.WithMaxRequests(1000),
        govisual.WithSharedStore(store),
    )
    
    // Start HTTP server
    log.Println("Starting dashboard server on :8080")
    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatalf("HTTP server error: %v", err)
    }
}
```

For NATS transport, you also need to add a NATS handler:

```go
// Set up NATS handler if using NATS transport
natsHandler, err := server.NewNATSHandler(store, "nats://localhost:4222")
if err != nil {
    log.Fatalf("Failed to create NATS handler: %v", err)
}

if err := natsHandler.Start(); err != nil {
    log.Fatalf("Failed to start NATS handler: %v", err)
}
defer natsHandler.Stop()
```

### 2. Setting Up gRPC Agents

Here's how to set up a gRPC agent with different transport options:

#### Shared Store Transport (Local Services)

```go
package main

import (
    "log"
    "net"
    
    "github.com/doganarif/govisual"
    "google.golang.org/grpc"
    "your-service/proto"
)

func main() {
    // Create or access shared store
    sharedStore, err := govisual.NewStore(
        govisual.WithMaxRequests(100),
        govisual.WithMemoryStorage(),
    )
    if err != nil {
        log.Fatalf("Failed to create store: %v", err)
    }
    
    // Create store transport
    transport := govisual.NewStoreTransport(sharedStore)
    
    // Create gRPC agent
    agent := govisual.NewGRPCAgent(transport,
        govisual.WithGRPCRequestDataLogging(true),
        govisual.WithGRPCResponseDataLogging(true),
    )
    
    // Create gRPC server with agent
    server := govisual.NewGRPCServer(agent)
    proto.RegisterYourServiceServer(server, &YourServiceImpl{})
    
    // Start server
    lis, err := net.Listen("tcp", ":9000")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    log.Println("Starting gRPC server on :9000")
    if err := server.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

#### HTTP Transport (Remote Services)

```go
// Create HTTP transport
transport := govisual.NewHTTPTransport("http://dashboard-server:8080/api/agent/logs",
    govisual.WithTimeout(5*time.Second),
    govisual.WithMaxRetries(3),
)

// Create gRPC agent
agent := govisual.NewGRPCAgent(transport,
    govisual.WithGRPCRequestDataLogging(true),
    govisual.WithGRPCResponseDataLogging(true),
    govisual.WithBatchingEnabled(true),
    govisual.WithBatchSize(10),
    govisual.WithBatchInterval(3*time.Second),
)
```

#### NATS Transport (Distributed Systems)

```go
// Create NATS transport
transport, err := govisual.NewNATSTransport("nats://nats-server:4222",
    govisual.WithMaxRetries(5),
    govisual.WithCredentials(map[string]string{
        "username": "user",
        "password": "pass",
    }),
)
if err != nil {
    log.Fatalf("Failed to create NATS transport: %v", err)
}

// Create gRPC agent
agent := govisual.NewGRPCAgent(transport,
    govisual.WithGRPCRequestDataLogging(true),
    govisual.WithGRPCResponseDataLogging(true),
)
```

### 3. Setting Up HTTP Agents

For HTTP services, use the HTTP agent:

```go
// Create transport 
transport := govisual.NewHTTPTransport("http://dashboard-server:8080/api/agent/logs")

// Create HTTP agent
agent := govisual.NewHTTPAgent(transport,
    govisual.WithHTTPRequestBodyLogging(true),
    govisual.WithHTTPResponseBodyLogging(true),
    govisual.WithMaxBodySize(1024*1024), // 1MB
    govisual.WithIgnorePaths("/health", "/metrics"),
    govisual.WithIgnoreExtensions(".jpg", ".png", ".css"),
)

// Apply as middleware to your HTTP server
mux := http.NewServeMux()
mux.HandleFunc("/", yourHandler)

// Wrap with agent middleware
http.ListenAndServe(":8000", agent.Middleware(mux))
```

## Configuration Options

### Agent Options

#### Common Options

```go
// Set maximum buffer size for when transport is unavailable
govisual.WithMaxBufferSize(100)

// Enable batching to reduce transport overhead
govisual.WithBatchingEnabled(true)

// Set batch size
govisual.WithBatchSize(20)

// Set batch interval
govisual.WithBatchInterval(5*time.Second)

// Add filtering to exclude certain requests
govisual.WithFilter(func(log *model.RequestLog) bool {
    // Skip health check endpoints
    if log.Type == model.TypeHTTP && log.Path == "/health" {
        return false
    }
    return true
})

// Add processing to modify or clean up logs before transport
govisual.WithProcessor(func(log *model.RequestLog) *model.RequestLog {
    // Redact sensitive information
    if log.Type == model.TypeHTTP && strings.Contains(log.Path, "/auth") {
        log.RequestBody = "[REDACTED]"
    }
    return log
})
```

#### gRPC Agent Options

```go
// Log request message data
govisual.WithGRPCRequestDataLogging(true)

// Log response message data
govisual.WithGRPCResponseDataLogging(true)

// Ignore specific gRPC methods
govisual.WithIgnoreGRPCMethods(
    "/health.HealthService/Check",
    "/grpc.reflection.v1.ReflectionService/*",
)
```

#### HTTP Agent Options

```go
// Log request bodies
govisual.WithHTTPRequestBodyLogging(true)

// Log response bodies
govisual.WithHTTPResponseBodyLogging(true)

// Set maximum body size to log
govisual.WithMaxBodySize(512*1024) // 512KB

// Ignore specific paths
govisual.WithIgnorePaths("/health", "/metrics", "/favicon.ico")

// Ignore specific file extensions
govisual.WithIgnoreExtensions(".jpg", ".png", ".gif", ".css", ".js")

// Transform paths before logging (e.g., to normalize UUIDs)
govisual.WithPathTransformer(func(path string) string {
    // Replace UUIDs with placeholders
    return regexp.MustCompile(`/users/[0-9a-f-]{36}`).
        ReplaceAllString(path, "/users/:id")
})
```

### Transport Options

```go
// Set endpoint for HTTP transport
govisual.WithEndpoint("http://dashboard-server:8080/api/agent/logs")

// Add authentication credentials
govisual.WithCredentials(map[string]string{
    "token": "your-auth-token",
    "api_key": "your-api-key",
})

// Configure retries
govisual.WithMaxRetries(5)
govisual.WithRetryWait(2*time.Second)

// Set timeout
govisual.WithTimeout(10*time.Second)

// Set buffer size for when transport is unavailable
govisual.WithBufferSize(200)
```

## Deployment Scenarios

### Single Service (Development)

For local development with a single service, use the shared store transport:

```
┌──────────────────┐
│  Single Service  │
│                  │
│ ┌─────────────┐  │
│ │  gRPC/HTTP  │  │
│ │   Agent     │  │
│ └─────┬───────┘  │
│       │          │
│ ┌─────▼───────┐  │
│ │   Shared    │  │
│ │    Store    │  │
│ └─────┬───────┘  │
│       │          │
│ ┌─────▼───────┐  │
│ │  Dashboard  │  │
│ └─────────────┘  │
└──────────────────┘
```

### Multiple Services (Single Machine)

For multiple services on a single machine, use the shared store transport:

```
┌──────────────────┐  ┌──────────────────┐
│  Service A (gRPC)│  │  Service B (HTTP)│
│                  │  │                  │
│ ┌─────────────┐  │  │ ┌─────────────┐  │
│ │ gRPC Agent  │  │  │ │ HTTP Agent  │  │
│ └──────┬──────┘  │  │ └──────┬──────┘  │
└────────┼─────────┘  └────────┼─────────┘
         │                     │
         ▼                     ▼
┌───────────────────────────────────────┐
│               Shared Store             │
└─────────────────┬─────────────────────┘
                  │
┌─────────────────▼─────────────────────┐
│            Dashboard Server            │
└───────────────────────────────────────┘
```

### Distributed Services (Multiple Machines)

For distributed services, use the HTTP or NATS transport:

```
┌──────────────────┐  ┌──────────────────┐
│  Service A (gRPC)│  │  Service B (HTTP)│
│                  │  │                  │
│ ┌─────────────┐  │  │ ┌─────────────┐  │
│ │ gRPC Agent  │  │  │ │ HTTP Agent  │  │
│ └──────┬──────┘  │  │ └──────┬──────┘  │
└────────┼─────────┘  └────────┼─────────┘
         │                     │
         │                     │
         ▼                     ▼
┌───────────────────────────────────────┐
│          Transport Layer              │
│      (NATS or HTTP Transport)         │
└─────────────────┬─────────────────────┘
                  │
                  │
┌─────────────────▼─────────────────────┐
│            Dashboard Server            │
│                                        │
│ ┌────────────────────────────────────┐ │
│ │           Store Backend            │ │
│ └────────────────────────────────────┘ │
└───────────────────────────────────────┘
```

## Best Practices

### Security Considerations

1. **Data Privacy**: Use the processor option to redact sensitive information from logs before transport:

```go
govisual.WithProcessor(func(log *model.RequestLog) *model.RequestLog {
    // Redact authentication tokens
    if log.Type == model.TypeHTTP {
        // Redact Authorization headers
        if auth, ok := log.RequestHeaders["Authorization"]; ok {
            log.RequestHeaders["Authorization"] = []string{"[REDACTED]"}
        }
        
        // Redact sensitive JSON fields in bodies
        if strings.Contains(log.Path, "/users") || strings.Contains(log.Path, "/accounts") {
            // Use a JSON parser to selectively redact fields
            if strings.Contains(log.RequestBody, "password") || 
               strings.Contains(log.RequestBody, "credit_card") {
                // Replace with redacted version
                log.RequestBody = redactSensitiveJSON(log.RequestBody)  
            }
        }
    }
    return log
})
```

2. **Transport Security**: Use secure connections for remote transports:

```go
// For HTTP transport
transport := govisual.NewHTTPTransport("https://dashboard-server:8080/api/agent/logs",
    govisual.WithCredentials(map[string]string{
        "token": "your-secure-token",
    }),
)

// For NATS transport with TLS
transport, err := govisual.NewNATSTransport("nats://nats-server:4222",
    govisual.WithCredentials(map[string]string{
        "token": "your-nats-token",
    }),
    // Add TLS configurations
)
```

### Performance Considerations

1. **Batching**: Enable batching to reduce network overhead for remote transports:

```go
agent := govisual.NewGRPCAgent(transport,
    govisual.WithBatchingEnabled(true),
    govisual.WithBatchSize(20),
    govisual.WithBatchInterval(5*time.Second),
)
```

2. **Filtering**: Filter out high-volume, low-value requests to reduce load:

```go
govisual.WithFilter(func(log *model.RequestLog) bool {
    // Skip static assets
    if log.Type == model.TypeHTTP {
        if strings.HasPrefix(log.Path, "/static/") || 
           strings.HasPrefix(log.Path, "/assets/") {
            return false
        }
    }
    
    // Skip health checks and metrics endpoints
    if log.Type == model.TypeGRPC && 
       (strings.Contains(log.GRPCService, "Health") || 
        strings.Contains(log.GRPCService, "Metrics")) {
        return false
    }
    
    return true
})
```

3. **Body Size Limits**: Limit the size of request/response bodies to prevent memory issues:

```go
// For HTTP agent
govisual.WithMaxBodySize(512*1024)  // 512KB limit

// For gRPC agent, process large messages
govisual.WithProcessor(func(log *model.RequestLog) *model.RequestLog {
    // Truncate large request/response data
    const maxSize = 1024 * 1024  // 1MB
    
    if log.Type == model.TypeGRPC {
        if len(log.GRPCRequestData) > maxSize {
            log.GRPCRequestData = log.GRPCRequestData[:maxSize] + "... [TRUNCATED]"
        }
        
        if len(log.GRPCResponseData) > maxSize {
            log.GRPCResponseData = log.GRPCResponseData[:maxSize] + "... [TRUNCATED]"
        }
    }
    
    return log
})
```

### Monitoring & Troubleshooting

Add logging to help debug agent and transport issues:

```go
// Create a processor that logs when errors occur
govisual.WithProcessor(func(log *model.RequestLog) *model.RequestLog {
    if log.Error != "" {
        internalLogger.Debugf("Request error captured: %s, path: %s", log.Error, log.Path)
    }
    
    // For gRPC status codes other than OK
    if log.Type == model.TypeGRPC && log.GRPCStatusCode != 0 {
        internalLogger.Debugf("gRPC request failed: status=%d, desc=%s, method=%s/%s",
            log.GRPCStatusCode, log.GRPCStatusDesc, log.GRPCService, log.GRPCMethod)
    }
    
    return log
})
```

## Advanced Use Cases

### Custom Transport Implementation

You can implement your own transport by implementing the `transport.Transport` interface:

```go
type CustomTransport struct {
    // Custom fields
}

func NewCustomTransport() *CustomTransport {
    return &CustomTransport{}
}

func (t *CustomTransport) Send(log *model.RequestLog) error {
    // Implement your custom sending logic
    return nil
}

func (t *CustomTransport) SendBatch(logs []*model.RequestLog) error {
    // Implement your custom batch sending logic
    return nil
}

func (t *CustomTransport) Close() error {
    // Clean up resources
    return nil
}
```

### Multiple Transport Targets

To send data to multiple visualization servers, create a composite transport:

```go
type CompositeTransport struct {
    transports []transport.Transport
}

func NewCompositeTransport(transports ...transport.Transport) *CompositeTransport {
    return &CompositeTransport{
        transports: transports,
    }
}

func (t *CompositeTransport) Send(log *model.RequestLog) error {
    var lastErr error
    for _, transport := range t.transports {
        if err := transport.Send(log); err != nil {
            lastErr = err
        }
    }
    return lastErr
}

func (t *CompositeTransport) SendBatch(logs []*model.RequestLog) error {
    var lastErr error
    for _, transport := range t.transports {
        if err := transport.SendBatch(logs); err != nil {
            lastErr = err
        }
    }
    return lastErr
}

func (t *CompositeTransport) Close() error {
    var lastErr error
    for _, transport := range t.transports {
        if err := transport.Close(); err != nil {
            lastErr = err
        }
    }
    return lastErr
}

// Usage
httpTransport := govisual.NewHTTPTransport("http://dashboard-1:8080/api/agent/logs")
natsTransport, _ := govisual.NewNATSTransport("nats://nats-server:4222")
compositeTransport := NewCompositeTransport(httpTransport, natsTransport)

agent := govisual.NewGRPCAgent(compositeTransport)
```

## Compatibility with Existing Code

The agent architecture is designed to be backward compatible with the existing GoVisual API. You can gradually migrate your codebase to use agents:

### Before (Direct Dashboard)

```go
grpcServer := grpc.NewServer(
    govisual.WrapGRPCServer(
        govisual.WithGRPC(true),
        govisual.WithGRPCRequestDataLogging(true),
        govisual.WithGRPCResponseDataLogging(true),
    )...)
```

### After (Agent Architecture)

```go
// Create transport
transport := govisual.NewStoreTransport(sharedStore)

// Create agent
agent := govisual.NewGRPCAgent(transport,
    govisual.WithGRPCRequestDataLogging(true),
    govisual.WithGRPCResponseDataLogging(true),
)

// Create server with agent
grpcServer := govisual.NewGRPCServer(agent)
```

## Conclusion

The agent architecture provides a flexible way to collect and visualize request data from distributed services. By separating data collection (agents) from visualization (dashboard), GoVisual can now monitor complex, multi-service architectures while providing options for different deployment scenarios.

Choose the right transport mechanism based on your deployment needs:

- **Shared Store**: For services running on the same machine
- **HTTP Transport**: For remote services with direct HTTP connectivity
- **NATS Transport**: For distributed systems where a message broker is available

The architecture is designed to be extensible, allowing for custom transport implementations and advanced use cases.
