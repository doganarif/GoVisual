# Storage Backend Options

GoVisual supports multiple storage backends for persisting request logs, allowing you to choose the option that best fits your needs.

## Available Storage Backends

### In-Memory Storage (Default)

The in-memory storage keeps all request logs in memory. This is the simplest option and requires no additional setup, but logs will be lost when the application restarts.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMemoryStorage(), // Optional, this is the default
)
```

**Pros:**

- No external dependencies
- Fast performance
- Zero configuration

**Cons:**

- Logs are lost on restart
- Limited by available memory
- Not suitable for long-term storage

### PostgreSQL Storage

For persistent storage of request logs, you can use PostgreSQL. This requires the `github.com/lib/pq` package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithPostgresStorage(
        "postgres://user:password@localhost:5432/dbname?sslmode=disable", // Connection string
        "govisual_requests"  // Table name (created automatically if it doesn't exist)
    ),
)
```

**Pros:**

- Persistent storage
- Logs retained across restarts
- SQL querying capabilities
- Reliable and mature storage

**Cons:**

- External dependency on PostgreSQL
- Requires database setup and maintenance
- Slightly higher latency than in-memory

**Schema:**

The PostgreSQL adapter automatically creates a table with the following schema:

```sql
CREATE TABLE IF NOT EXISTS govisual_requests (
    id TEXT PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE,
    method TEXT,
    path TEXT,
    query TEXT,
    request_headers JSONB,
    response_headers JSONB,
    status_code INTEGER,
    duration BIGINT,
    request_body TEXT,
    response_body TEXT,
    error TEXT,
    middleware_trace JSONB,
    route_trace JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
)
```

### Redis Storage

For high-performance storage with automatic expiration capabilities, you can use Redis. This requires the `github.com/go-redis/redis/v8` package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithRedisStorage(
        "redis://localhost:6379/0", // Redis connection string
        86400                       // TTL in seconds (24 hours)
    ),
)
```

**Pros:**

- Fast performance
- Automatic time-to-live (TTL) support
- Persistence options (RDB/AOF)
- Smaller memory footprint than in-memory store

**Cons:**

- External dependency on Redis
- Requires setup and maintenance
- Less querying capabilities than SQL

**Storage Structure:**

The Redis adapter uses the following storage structure:

- Each request log is stored as a JSON string with key `govisual:{id}`
- A sorted set named `govisual:logs` is used to maintain order by timestamp
- All keys automatically expire based on the configured TTL

### SQLite Storage

For lightweight, persistent local storage, you can use SQLite. This requires the github.com/ncruces/go-sqlite3 package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithSQLiteStorage(
        "./govisual.db",      // Path to the SQLite database file
        "govisual_requests",  // Table name (created automatically if it doesn't exist)
    ),
)
```

**Pros:**

- Persistent local storage with no external server required
- Zero configuration: just a .db file
- Great for development, testing, and embedded environments
- SQL querying capabilities

**Cons:**

- Not recommended for high concurrency or large-scale production use
- Less scalable than PostgreSQL or Redis
- Database file can grow quickly under heavy usage

**Schema:**

The SQLite adapter automatically creates a table with the following schema:

```sql
CREATE TABLE IF NOT EXISTS govisual_requests (
    id TEXT PRIMARY KEY,
    timestamp DATETIME,
    method TEXT,
    path TEXT,
    query TEXT,
    request_headers TEXT,
    response_headers TEXT,
    status_code INTEGER,
    duration INTEGER,
    request_body TEXT,
    response_body TEXT,
    error TEXT,
    middleware_trace TEXT,
    route_trace TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

**Summary of usage:**

- **For local persistence and simplicity**: SQLite is a great choice.
- **For environments without external dependencies**: Just point to a .db file and use.

### SQLite with Existing Connection

If your application already uses SQLite with a driver like `github.com/mattn/go-sqlite3`, you may experience a conflict when using govisual's SQLite storage due to multiple driver registrations. To avoid this conflict, you can pass your existing database connection to govisual:

```go
import (
    "database/sql"

    "github.com/doganarif/govisual"
    _ "github.com/mattn/go-sqlite3" // Your preferred SQLite driver
)

func main() {
    // Create your own database connection
    db, err := sql.Open("sqlite3", "./govisual.db")
    if err != nil {
        // Handle error
    }
    defer db.Close()

    // Pass the existing connection to govisual
    handler := govisual.Wrap(
        mux,
        govisual.WithSQLiteStorageDB(db, "govisual_requests"),
    )

    // Use the handler
    http.ListenAndServe(":8080", handler)
}
```

**Pros:**

- Avoids "sql: Register called twice for driver sqlite3" panic
- Allows you to use your preferred SQLite driver
- Compatible with any existing SQLite connection
- Share a single connection across your application

**Cons:**

- Requires you to manage the database connection lifecycle
- You need to ensure your driver is compatible with govisual's requirements

**When to use:**

- When you're already using SQLite elsewhere in your application
- When you want to avoid driver registration conflicts
- When you need more control over the database connection

### MongoDB Storage

For document-based storage with high scalability, you can use MongoDB. This requires the `go.mongodb.org/mongo-driver/v2/mongo` package.

```go
handler := govisual.Wrap(
    mux,
    govisual.WithMongoDBStorage(
        "mongodb://user:password@localhost:27017",  // MongoDB connection string
        "govisual",                                 // Database name
        "requests"                                  // Collection name
    ),
)
```

**Pros:**

- Schema flexibility for varying request/response structures
- High performance for read/write operations
- Horizontal scalability through sharding
- Rich querying capabilities for request analysis
- Native JSON support for HTTP request/response data
- Efficient handling of large document sizes
- Built-in replication and high availability

**Cons:**

- Higher memory usage compared to SQL databases
- Requires separate MongoDB instance setup and maintenance
- Eventually consistent by default (may affect real-time tracking)
- Steeper learning curve for team operations
- Higher hosting costs for production deployments
- More complex backup procedures

**Storage Structure:**

The MongoDB adapter stores request logs in the following structure:

```json
{
    "_id": "unique_request_id",
    "timestamp": ISODate("2024-03-21T10:00:00Z"),
    "method": "GET",
    "path": "/api/users",
    "query": "?page=1",
    "request_headers": {
        "Content-Type": "application/json",
        "Authorization": "Bearer ..."
    },
    "response_headers": {
        "Content-Type": "application/json"
    },
    "status_code": 200,
    "duration": 150,
    "request_body": "...",
    "response_body": "...",
    "error": null,
    "middleware_trace": [...],
    "route_trace": [...],
    "created_at": ISODate("2024-03-21T10:00:00Z")
}
```

**Connection String Format:**

Standard MongoDB connection string format:
```
mongodb://[username:password@]host[:port][/database][?options]
```

Example:
```
mongodb://admin:password@localhost:27017/govisual?authSource=admin
```

**When to use:**

- For applications with varying request/response structures
- When high scalability is a requirement
- In microservices architectures where eventual consistency is acceptable
- When native JSON storage and querying is important
- For distributed systems requiring horizontal scaling

## Choosing a Storage Backend

Here are some guidelines for choosing the appropriate storage backend:

1. **Development/Testing:**

   - In-memory storage (default) is typically sufficient
   - No setup required, just works out of the box

2. **Production/Longer-term storage:**

   - PostgreSQL for permanent storage with SQL querying capabilities
   - Redis for high-performance with TTL-based cleanup

3. **High-traffic applications:**
   - Redis for high throughput and lower memory footprint
   - PostgreSQL with proper indexing for long-term storage and analytics

## Connection String Formats

### PostgreSQL

Standard PostgreSQL connection string format:

```
postgres://[username]:[password]@[host]:[port]/[database_name]?[parameters]
```

Example:

```
postgres://postgres:password@localhost:5432/govisual?sslmode=disable
```

### Redis

Standard Redis connection string format:

```
redis://[username]:[password]@[host]:[port]/[database_number]
```

Example:

```
redis://user:password@localhost:6379/0
```

## Graceful Shutdown

GoVisual automatically handles graceful shutdown for all storage backends. When the application receives a shutdown signal (SIGTERM, SIGINT), it will properly close database connections.

## Example

For a complete example of all storage backends, see the [Multi-Storage Example](../cmd/examples/multistorage/README.md).