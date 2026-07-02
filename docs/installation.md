# Installation

This guide covers installing GoVisual for your Go web applications.

## Requirements

- Go 1.24 or higher
- (Optional) a database for persistent storage — each backend is its own module under `store/`
- (Optional) an OpenTelemetry collector, via the `telemetry` module

## Using Go Modules (Recommended)

The simplest way to install GoVisual is via Go modules:

```bash
go get github.com/doganarif/govisual/v2
```

The core module has no database drivers or gRPC. Add-ons install separately, only when you use them:

```bash
go get github.com/doganarif/govisual/store/postgres   # or redis, sqlite, mongodb
go get github.com/doganarif/govisual/telemetry        # OpenTelemetry export
go get github.com/doganarif/govisual/mcp              # MCP server for coding agents
```

## Verifying Installation

Create a simple test application to verify that GoVisual is working correctly:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/doganarif/govisual/v2"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, GoVisual!")
    })

    // Wrap with GoVisual
    handler := govisual.Wrap(mux)

    fmt.Println("Server starting at http://localhost:8080")
    fmt.Println("GoVisual dashboard available at http://localhost:8080/__viz")
    http.ListenAndServe(":8080", handler)
}
```

If everything is working correctly, you should be able to:

1. Access your application at http://localhost:8080/
2. See the GoVisual dashboard at http://localhost:8080/\_\_viz

## Next Steps

Once GoVisual is installed, check out the [Quick Start Guide](quick-start.md) to learn how to use it in your application.
