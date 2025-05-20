# GoVisual Multi-Storage Example

This example demonstrates how to use GoVisual with different storage backends:

- In-memory storage (default)
- PostgreSQL
- Redis

## Running the Example

You can run the example using Docker Compose, which will set up all necessary services:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

By default, the application will use in-memory storage. You can change the storage backend by modifying the environment variables in `docker-compose.yml`:

### Using In-Memory Storage (Default)

```yaml
environment:
  - PORT=8080
  # No additional environment variables needed for in-memory storage
```

### Using PostgreSQL Storage

```yaml
environment:
  - PORT=8080
  - GOVISUAL_STORAGE_TYPE=postgres
  - GOVISUAL_PG_CONN=postgres://postgres:postgres@postgres:5432/govisual?sslmode=disable
  - GOVISUAL_PG_TABLE=govisual_requests
```

### Using Redis Storage

```yaml
environment:
  - PORT=8080
  - GOVISUAL_STORAGE_TYPE=redis
  - GOVISUAL_REDIS_CONN=redis://redis:6379/0
  - GOVISUAL_REDIS_TTL=86400
```

### Using SQLite Storage

```yaml
environment:
  - PORT=8080
  - GOVISUAL_STORAGE_TYPE=sqlite
  - GOVISUAL_SQLITE_DBPATH=/data/govisual.db
  - GOVISUAL_SQLITE_TABLE=govisual_requests
```

### Using MongoDB Storage

```yaml
GOVISUAL_STORAGE_TYPE=mongodb
GOVISUAL_MONGO_URI=mongodb://root:root@localhost:27017/
GOVISUAL_MONGO_DATABASE=logs
GOVISUAL_MONGO_COLLECTION=request_logs
```

## Accessing the Application

Once the application is running, you can access it at:

- Main application: http://localhost:8080
- GoVisual dashboard: http://localhost:8080/\_\_viz

## Available Endpoints

- `GET /`: Home page with instructions
- `GET /api/users`: List users (JSON)
- `POST /api/users`: Create user (expects JSON body)
- `GET /api/products`: List products (JSON)
- `POST /api/products`: Create product (expects JSON body)

## Running Without Docker

If you prefer to run the example without Docker, you can use:

```bash
# In-memory storage (default)
go run main.go

# PostgreSQL storage
GOVISUAL_STORAGE_TYPE=postgres \
GOVISUAL_PG_CONN="postgres://postgres:postgres@localhost:5432/govisual?sslmode=disable" \
go run main.go

# Redis storage
GOVISUAL_STORAGE_TYPE=redis \
GOVISUAL_REDIS_CONN="redis://localhost:6379/0" \
go run main.go
```
