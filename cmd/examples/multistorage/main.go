package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/doganarif/govisual"
)

func main() {
	// Default storage is in-memory, but can be changed via environment variables
	var opts []govisual.Option

	// Add basic options
	opts = append(opts,
		govisual.WithMaxRequests(100),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	// Check for storage configuration
	storageType := os.Getenv("GOVISUAL_STORAGE_TYPE")

	switch storageType {
	case "postgres":
		connStr := os.Getenv("GOVISUAL_PG_CONN")
		if connStr == "" {
			log.Fatal("PostgreSQL connection string not provided in GOVISUAL_PG_CONN")
		}
		tableName := os.Getenv("GOVISUAL_PG_TABLE")
		if tableName == "" {
			tableName = "govisual_requests"
		}
		opts = append(opts, govisual.WithPostgresStorage(connStr, tableName))
		log.Printf("Using PostgreSQL storage with table: %s", tableName)

	case "redis":
		connStr := os.Getenv("GOVISUAL_REDIS_CONN")
		if connStr == "" {
			log.Fatal("Redis connection string not provided in GOVISUAL_REDIS_CONN")
		}
		ttl := 86400 // 24 hours by default
		if ttlStr := os.Getenv("GOVISUAL_REDIS_TTL"); ttlStr != "" {
			var err error
			ttl, err = parseInt(ttlStr)
			if err != nil {
				log.Printf("Invalid TTL value: %s, using default of 86400 seconds", ttlStr)
				ttl = 86400
			}
		}
		opts = append(opts, govisual.WithRedisStorage(connStr, ttl))
		log.Printf("Using Redis storage with TTL: %d seconds", ttl)

	case "sqlite":
		connStr := os.Getenv("GOVISUAL_SQLITE_DBPATH")
		if connStr == "" {
			log.Fatal("SQLite database path not provided in GOVISUAL_SQLITE_DBPATH")
		}

		tableName := os.Getenv("GOVISUAL_SQLITE_TABLE")
		if tableName == "" {
			tableName = "govisual_requests"
		}

		opts = append(opts, govisual.WithSQLiteStorage(connStr, tableName))
		log.Printf("Using SQLite storage with table: %s", tableName)

	default:
		// Default to memory storage
		opts = append(opts, govisual.WithMemoryStorage())
		log.Println("Using in-memory storage (default)")
	}

	// Create a simple HTTP handler
	mux := http.NewServeMux()

	// Add some example routes
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/users", usersHandler)
	mux.HandleFunc("/api/products", productsHandler)

	// Wrap with GoVisual
	handler := govisual.Wrap(mux, opts...)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Access the dashboard at http://localhost:%s/__viz", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to GoVisual Multi-Storage Example!\n\n"))
	w.Write([]byte("Available endpoints:\n"))
	w.Write([]byte("- GET /api/users: List users\n"))
	w.Write([]byte("- POST /api/users: Create user (with JSON body)\n"))
	w.Write([]byte("- GET /api/products: List products\n"))
	w.Write([]byte("- POST /api/products: Create product (with JSON body)\n"))
	w.Write([]byte("\nAccess the dashboard at /__viz\n"))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing time
	time.Sleep(50 * time.Millisecond)

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`))

	case http.MethodPost:
		// Simulate request processing
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 3, "name": "New User", "created": true}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing time
	time.Sleep(75 * time.Millisecond)

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"products": [{"id": 1, "name": "Laptop"}, {"id": 2, "name": "Phone"}]}`))

	case http.MethodPost:
		// Simulate request processing
		time.Sleep(125 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 3, "name": "New Product", "created": true}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
