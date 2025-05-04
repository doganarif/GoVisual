# Installation

This guide covers installing GoVisual for your Go web applications.

## Requirements

- Go 1.19 or higher
- (Optional) PostgreSQL for persistent storage backend
- (Optional) Redis for high-performance storage backend
- (Optional) OpenTelemetry collector for telemetry data export

## Using Go Modules (Recommended)

The simplest way to install GoVisual is via Go modules:

```bash
go get github.com/doganarif/govisual
```

## Manual Installation

You can also manually clone the repository:

```bash
git clone https://github.com/doganarif/govisual.git
cd govisual
go install
```

## Verifying Installation

Create a simple test application to verify that GoVisual is working correctly:

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/doganarif/govisual"
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
