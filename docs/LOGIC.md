# GoVisual: Core Logic

This document outlines the core logic and architecture of GoVisual, a request visualization tool for Go HTTP applications.

## Overview

GoVisual is designed to be a lightweight, non-intrusive tool that wraps your existing HTTP handlers to provide visibility into request processing. The fundamental design principles include:

1. **Middleware Architecture**: GoVisual uses the HTTP middleware pattern to intercept requests before and after your handler processes them
2. **In-Memory Storage**: All request data is stored in a circular buffer with configurable size
3. **Contextual Enrichment**: The middleware captures detailed execution metrics including timing data
4. **Dashboard Rendering**: A self-contained HTML dashboard for visualizing captured requests
5. **Zero External Dependencies**: No third-party packages required beyond the Go standard library
6. **OpenTelemetry Integration**: Optional integration with OpenTelemetry for distributed tracing

## Request Flow

When a request is processed through GoVisual:

1. The request is intercepted by the GoVisual middleware
2. Request metadata (headers, path, method, body) is captured
3. The original handler is called to process the request
4. Response metadata (status code, headers, body, timing) is captured
5. Both are stored in the in-memory circular buffer
6. The dashboard path (`/__viz` by default) is intercepted to serve the visualization UI
7. If OpenTelemetry is enabled, a trace is created and exported to the configured endpoint

## Core Components

The diagram below illustrates the relationship between core components:

```mermaid
flowchart TD
    subgraph "Developer Workflow"
        A1[Run Local App with GoVisual] --> A2["Open Browser: localhost:8080/__viz"]
        A2 --> A3[Monitor Requests in Dashboard]
        A3 --> A4[Debug Issues & Optimize App]
    end

    subgraph "HTTP Request Flow"
        B1[Local API Request] --> B2[govisual.Wrap]
        B2 --> B3{Is Dashboard Path?}
        B3 -->|Yes| B4[Serve GoVisual Dashboard]
        B3 -->|No| B5[Process Through Middleware]
        B5 --> B6[Forward to App Handler]
        B6 --> B7[Return to User]
    end

    subgraph "Visualization & Debug Features"
        C1[Real-time Request Table] --> C2[Filter & Search]
        C1 --> C3[Detailed Request View]
        C3 --> C4[Headers Inspection]
        C3 --> C5[Request Body Viewer]
        C3 --> C6[Response Body Viewer]
        C3 --> C7[Status & Timing Info]
        C1 --> C8[Middleware Trace Visualization]
        C8 --> C9[Performance Bottleneck Detection]
    end

    subgraph "Data Capture System"
        D1[Intercept HTTP Request] --> D2[Capture Headers]
        D1 --> D3[Capture Request Body]
        D1 --> D4[Measure Start Time]
        D5[Wrap Response Writer] --> D6[Capture Status Code]
        D5 --> D7[Capture Response Body]
        D5 --> D8[Calculate Duration]
        D9[Track Middleware Flow] --> D10[Record Execution Times]
    end

    subgraph "Development Configuration"
        E1[Feature Flags] --> E2[Enable in Dev Only]
        E1 --> E3[Disable in Production]
        E4[Logging Options] --> E5[Request Body Capture]
        E4 --> E6[Response Body Capture]
        E7[Storage Options] --> E8[Memory Limits]
    end

    subgraph "OpenTelemetry Integration"
        F1[Initialize OTel] --> F2[Create Traces]
        F2 --> F3[Add HTTP Attributes]
        F3 --> F4[Export to Backend]
        F4 --> F5[View in Jaeger/Other UI]
    end

    B7 --> A3
    D10 --> C8
    D8 --> C7
    D3 --> C5
    D7 --> C6
    E2 --> A1
    B5 --> F2
```

## Technical Architecture

The following class diagram shows the relationships between key components:

```mermaid
classDiagram
    class Config {
        +int MaxRequests
        +string DashboardPath
        +bool LogRequestBody
        +bool LogResponseBody
        +string[] IgnorePaths
        +bool EnableOpenTelemetry
        +string ServiceName
        +string ServiceVersion
        +string OTelEndpoint
        +ShouldIgnorePath(path string) bool
    }

    class Wrap {
        +Wrap(handler http.Handler, opts) http.Handler
    }

    class Option {
        +func(c *Config)
    }

    class RequestLog {
        +string ID
        +time.Time Timestamp
        +string Method
        +string Path
        +string Query
        +http.Header RequestHeaders
        +http.Header ResponseHeaders
        +int StatusCode
        +int64 Duration
        +string RequestBody
        +string ResponseBody
        +string Error
        +object MiddlewareTrace
        +object RouteTrace
    }

    class Store {
        +Add(log *RequestLog)
        +Get(id string)
        +GetAll()
        +Clear()
        +GetLatest(n int)
    }

    class InMemoryStore {
        -logs
        -int capacity
        -int size
        -int next
        -mutex mu
        +Add(log *RequestLog)
        +Get(id string)
        +GetAll()
        +Clear()
        +GetLatest(n int)
    }

    class Handler {
        +Store store
        +Profiler profiler
        +staticFS
        +ServeHTTP(w, r)
        +handleAPIRequests(w, r)
        +handleClearRequests(w, r)
        +handleSSE(w, r)
        +handleCompareRequests(w, r)
        +handleReplayRequest(w, r)
        +handleMetrics(w, r)
        +handleFlameGraph(w, r)
        +handleBottlenecks(w, r)
        +handleSystemInfo(w, r)
    }

    class Middleware {
        +Wrap(handler, store, logRequest, logResponse, matcher)
    }

    class PathMatcher {
        +ShouldIgnorePath(path string) bool
    }

    class OTelMiddleware {
        +tracer trace.Tracer
        +propagator propagation.TextMapPropagator
        +handler http.Handler
        +serviceVersion string
        +ServeHTTP(w, r)
    }

    class TelemetryInit {
        +InitTracer(ctx, serviceName, serviceVersion, endpoint)
    }

    Config ..|> PathMatcher
    Wrap --> Config
    Option --> Config
    InMemoryStore ..|> Store
    Middleware --> Store
    Middleware --> PathMatcher
    Handler --> Store
    Wrap --> Handler
    Wrap --> Middleware
    Middleware --> RequestLog
    Config --> OTelMiddleware
    Config --> TelemetryInit
    Wrap --> OTelMiddleware
```

## Implementation Details

### Configuration Options

GoVisual can be configured with various options:

- **MaxRequests**: Maximum number of requests to store in memory (default: 100)
- **DashboardPath**: Path to access the dashboard (default: "/\_\_viz")
- **LogRequestBody**: Enable request body logging (default: false)
- **LogResponseBody**: Enable response body logging (default: false)
- **IgnorePaths**: Paths to ignore (default: empty)
- **EnableOpenTelemetry**: Enable OpenTelemetry integration (default: false)
- **ServiceName**: Service name for OpenTelemetry (default: "govisual")
- **ServiceVersion**: Service version for OpenTelemetry (default: "dev")
- **OTelEndpoint**: OTLP exporter endpoint (default: "localhost:4317")

### Circular Buffer Implementation

The in-memory storage uses a circular buffer implementation for efficient memory usage:

1. A fixed-size array holds request logs
2. New logs replace old ones when the buffer is full
3. A pointer tracks the next insertion position
4. A size counter tracks the number of valid entries

This ensures memory usage remains constant regardless of request volume.

### Dashboard Implementation

The dashboard is implemented using a React-based frontend with a Go backend API:

1. React frontend built with Preact and TypeScript, bundled and embedded using Go's `embed` package
2. Server-Sent Events (SSE) provide real-time updates
3. Client-side filtering, sorting, and visualization minimize server load
4. RESTful API endpoints serve JSON data for requests, metrics, and system information
5. Advanced features include flame graphs, performance profiling, and request comparison

### OpenTelemetry Integration

When enabled, GoVisual integrates with OpenTelemetry to provide distributed tracing:

1. Initialize a tracer provider and exporter on startup
2. Create a middleware to wrap HTTP handlers and create spans
3. Add relevant HTTP attributes to spans (method, path, status, etc.)
4. Export traces to the configured endpoint (typically Jaeger or OpenTelemetry Collector)
5. Provide context propagation for distributed tracing

The integration is designed to be optional and non-intrusive, allowing users to benefit from both GoVisual's dashboard and OpenTelemetry's ecosystem.
