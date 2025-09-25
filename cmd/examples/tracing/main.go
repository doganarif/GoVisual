package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/doganarif/govisual"
	"github.com/doganarif/govisual/internal/middleware"
	_ "github.com/mattn/go-sqlite3"
)

// Example middleware that adds to the trace
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get tracer from context
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.StartTrace("Logging Middleware", "middleware", map[string]interface{}{
				"user_agent":  r.UserAgent(),
				"remote_addr": r.RemoteAddr,
			})
			defer tracer.EndTrace(nil)
		}

		log.Printf("[%s] %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Auth middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.StartTrace("Authentication", "middleware", map[string]interface{}{
				"auth_type": "bearer",
			})
			defer tracer.EndTrace(nil)
		}

		// Simulate auth check
		time.Sleep(10 * time.Millisecond)

		// Check for auth header
		if auth := r.Header.Get("Authorization"); auth == "" {
			tracer.RecordCustom("Auth Failed", map[string]interface{}{
				"reason": "missing_header",
			})
		}

		next.ServeHTTP(w, r)
	})
}

// Database helper
type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create sample table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			email TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	// Insert sample data
	_, err = db.Exec(`
		INSERT INTO users (name, email) VALUES 
		('John Doe', 'john@example.com'),
		('Jane Smith', 'jane@example.com'),
		('Bob Wilson', 'bob@example.com')
	`)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) GetUsers(ctx context.Context) ([]map[string]interface{}, error) {
	tracer := middleware.GetTracer(ctx)
	if tracer != nil {
		tracer.StartTrace("Database Query", "sql", nil)
		defer tracer.EndTrace(nil)
	}

	query := "SELECT id, name, email, created_at FROM users"
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email, createdAt string
		if err := rows.Scan(&id, &name, &email, &createdAt); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"name":       name,
			"email":      email,
			"created_at": createdAt,
		})
	}

	return users, nil
}

func main() {
	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create main handler
	mux := http.NewServeMux()

	// Home route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.RecordCustom("Processing Request", map[string]interface{}{
				"handler": "home",
			})
		}

		response := map[string]interface{}{
			"message":   "Welcome to GoVisual Tracing Example",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Users route with database query
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.StartTrace("Get Users Handler", "handler", nil)
			defer tracer.EndTrace(nil)
		}

		// Simulate processing
		time.Sleep(20 * time.Millisecond)

		users, err := db.GetUsers(r.Context())
		if err != nil {
			if tracer != nil {
				tracer.RecordCustom("Database Error", map[string]interface{}{
					"error": err.Error(),
				})
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	})

	// Slow endpoint for testing
	mux.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.StartTrace("Slow Handler", "handler", nil)
			defer tracer.EndTrace(nil)
		}

		// Multiple operations
		if tracer != nil {
			tracer.StartTrace("Step 1: Validation", "custom", nil)
		}
		time.Sleep(50 * time.Millisecond)
		if tracer != nil {
			tracer.EndTrace(nil)
		}

		if tracer != nil {
			tracer.StartTrace("Step 2: Processing", "custom", nil)
		}
		time.Sleep(100 * time.Millisecond)
		if tracer != nil {
			tracer.EndTrace(nil)
		}

		if tracer != nil {
			tracer.StartTrace("Step 3: Formatting", "custom", nil)
		}
		time.Sleep(30 * time.Millisecond)
		if tracer != nil {
			tracer.EndTrace(nil)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Slow operation completed",
			"duration": "180ms",
		})
	})

	// External API call example
	mux.HandleFunc("/api/external", func(w http.ResponseWriter, r *http.Request) {
		tracer := middleware.GetTracer(r.Context())
		if tracer != nil {
			tracer.StartTrace("External API Handler", "handler", nil)
			defer tracer.EndTrace(nil)
		}

		// Simulate external HTTP call
		if tracer != nil {
			tracer.RecordHTTP("GET", "https://api.example.com/data",
				250*time.Millisecond, 200, nil)
		}
		time.Sleep(250 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "External data fetched",
			"source":  "api.example.com",
		})
	})

	// Apply middleware chain
	handler := http.Handler(mux)
	handler = authMiddleware(handler)
	handler = loggingMiddleware(handler)

	// Wrap with GoVisual middleware
	visualHandler := govisual.Wrap(handler,
		govisual.WithDashboardPath("/__viz"),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithProfiling(true),
		govisual.WithProfileThreshold(1*time.Millisecond),
	)

	// Start server
	fmt.Println("Server starting on http://localhost:8090")
	fmt.Println("Dashboard available at http://localhost:8090/__viz")
	fmt.Println("")
	fmt.Println("Try these endpoints to see tracing:")
	fmt.Println("  - http://localhost:8090/")
	fmt.Println("  - http://localhost:8090/api/users")
	fmt.Println("  - http://localhost:8090/api/slow")
	fmt.Println("  - http://localhost:8090/api/external")
	fmt.Println("")
	fmt.Println("Then check the Trace tab in the dashboard!")

	log.Fatal(http.ListenAndServe(":8090", visualHandler))
}
