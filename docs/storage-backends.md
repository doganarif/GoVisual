# Storage Backends

GoVisual's core module (`github.com/doganarif/govisual/v2`) has no database drivers. Each storage backend lives in its own Go module under `store/`. Pull in only what you use.

All backends implement `store.Store` and are passed to govisual via `WithStore`:

```go
import "github.com/doganarif/govisual/v2"

handler := govisual.Wrap(mux, govisual.WithStore(myStore))
```

## In-Memory (Default)

No install required. When no `WithStore` option is given, govisual creates an in-memory ring buffer bounded by `WithMaxRequests` (default 100).

You can also create one explicitly and share it — for example, to expose it to the MCP server:

```go
import "github.com/doganarif/govisual/v2/store"

st := store.NewMemory(500) // capacity
handler := govisual.Wrap(mux, govisual.WithStore(st))
```

Logs are lost on restart. Suitable for development.

## PostgreSQL

```bash
go get github.com/doganarif/govisual/store/postgres
```

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/postgres"
)

pg, err := postgres.New(
    "postgres://user:password@localhost:5432/dbname?sslmode=disable", // connection string
    "govisual_requests",                                               // table name
    500,                                                               // capacity (rows kept)
)
if err != nil {
    log.Fatal(err)
}
defer pg.Close()

handler := govisual.Wrap(mux, govisual.WithStore(pg))
```

The table is created automatically on first use. The capacity limit is enforced by a periodic trim: once every 32 inserts, rows beyond the limit are deleted. The `github.com/lib/pq` driver is included transitively.

**Schema:**

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

## Redis

```bash
go get github.com/doganarif/govisual/store/redis
```

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/redis"
)

rdb, err := redis.New(
    "redis://localhost:6379/0", // connection string
    500,                        // capacity (number of entries kept)
    86400,                      // TTL in seconds (24 hours; 0 = default 24h)
)
if err != nil {
    log.Fatal(err)
}
defer rdb.Close()

handler := govisual.Wrap(mux, govisual.WithStore(rdb))
```

Each request is stored as a JSON string keyed by `govisual:{id}`. A sorted set named `govisual:logs` maintains order by timestamp. Keys expire automatically per the configured TTL. The `github.com/go-redis/redis/v8` client is included transitively.

## SQLite

```bash
go get github.com/doganarif/govisual/store/sqlite
```

The SQLite module does not import a driver — it calls `sql.Open("sqlite3", path)` and expects the caller to have one registered. Import your preferred driver before calling `New`:

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/sqlite"
    _ "github.com/ncruces/go-sqlite3/driver"
    _ "github.com/ncruces/go-sqlite3/embed"
)

sq, err := sqlite.New(
    "./govisual.db",     // path to the database file
    "govisual_requests", // table name
    500,                 // capacity
)
if err != nil {
    log.Fatal(err)
}
defer sq.Close()

handler := govisual.Wrap(mux, govisual.WithStore(sq))
```

### Reusing an Existing Connection

If your application already opens a SQLite database, pass that `*sql.DB` to avoid double driver registration:

```go
import (
    "database/sql"
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/sqlite"
    _ "github.com/mattn/go-sqlite3"
)

db, err := sql.Open("sqlite3", "./app.db")
if err != nil {
    log.Fatal(err)
}

sq, err := sqlite.NewWithDB(db, "govisual_requests", 500)
if err != nil {
    log.Fatal(err)
}
// govisual does not close db when NewWithDB is used — you manage the lifecycle.

handler := govisual.Wrap(mux, govisual.WithStore(sq))
```

**Schema:**

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

## MongoDB

```bash
go get github.com/doganarif/govisual/store/mongodb
```

```go
import (
    "github.com/doganarif/govisual/v2"
    "github.com/doganarif/govisual/store/mongodb"
)

mdb, err := mongodb.New(
    "mongodb://user:password@localhost:27017", // URI
    "govisual",                                // database name
    "requests",                                // collection name
    500,                                       // capacity
)
if err != nil {
    log.Fatal(err)
}
defer mdb.Close()

handler := govisual.Wrap(mux, govisual.WithStore(mdb))
```

A descending index on `timestamp` is created automatically. The `go.mongodb.org/mongo-driver/v2` client is included transitively.

## Choosing a Backend

| Backend | Persists across restarts | External server | Notes |
| --- | --- | --- | --- |
| Memory | No | No | Default; zero setup |
| SQLite | Yes | No | Good for single-process dev/test |
| PostgreSQL | Yes | Yes | Use for long-term storage with SQL queries |
| Redis | Yes (with AOF/RDB) | Yes | Use for high throughput; entries expire by TTL |
| MongoDB | Yes | Yes | Use when you already run Mongo |

## Change Notification

`store.WithNotify` wraps any `store.Store` and signals subscribers after each `Add`. The dashboard uses it for live updates without polling:

```go
import "github.com/doganarif/govisual/v2/store"

st := store.WithNotify(store.NewMemory(500))

ch, cancel := st.Subscribe()
defer cancel()

go func() {
    for range ch {
        // a new request was stored
    }
}()
```

## Graceful Shutdown

Pass a context via `WithShutdownContext`. When the context is cancelled, govisual calls `Close()` on the store:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

handler := govisual.Wrap(
    mux,
    govisual.WithStore(pg),
    govisual.WithShutdownContext(ctx),
)
```

## Related Documentation

- [Configuration Options](configuration.md) - Full options reference
- [API Reference](api-reference.md) - store package API
