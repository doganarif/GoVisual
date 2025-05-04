# GoVisual OpenTelemetry Example

This example demonstrates GoVisual integration with OpenTelemetry.

## Prerequisites

- Docker and Docker Compose for running Jaeger

## Setup

1. Start Jaeger:

```bash
docker-compose up -d
```

2. Run the example:

```bash
go run main.go
```

This will start a server on port 8080 with OpenTelemetry instrumentation enabled.

## Endpoints

- `/` - Home page with links to API endpoints
- `/api/users` - Returns user data with nested spans
- `/api/search?q=test` - Search endpoint with query parameter as a span attribute
- `/api/health` - Health check endpoint (not traced due to ignore path)

## Accessing the Dashboards

- GoVisual Dashboard: [http://localhost:8080/\_\_viz](http://localhost:8080/__viz)
- Jaeger UI: [http://localhost:16686](http://localhost:16686)

## Features Demonstrated

- GoVisual integration with OpenTelemetry
- Creating custom spans and nested spans
- Adding attributes to spans
- Path-based span filtering
- Viewing traces in Jaeger
- Correlating requests between GoVisual and Jaeger
