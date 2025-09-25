package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/doganarif/govisual"
	"github.com/doganarif/govisual/internal/profiling"
	_ "github.com/mattn/go-sqlite3"
)

// Example structs
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// Database connection (for demo purposes)
var db *sql.DB

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.Parse()

	// Initialize database
	initDatabase()
	defer db.Close()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Add routes with varying performance characteristics
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/fast", fastHandler)
	mux.HandleFunc("/api/slow", slowHandler)
	mux.HandleFunc("/api/cpu-intensive", cpuIntensiveHandler)
	mux.HandleFunc("/api/memory-intensive", memoryIntensiveHandler)
	mux.HandleFunc("/api/database", databaseHandler)
	mux.HandleFunc("/api/external", externalAPIHandler)
	mux.HandleFunc("/api/complex", complexHandler)

	// Configure GoVisual with performance profiling
	handler := govisual.Wrap(mux,
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithProfiling(true),                      // Enable profiling
		govisual.WithProfileType(profiling.ProfileAll),    // Profile everything
		govisual.WithProfileThreshold(5*time.Millisecond), // Profile requests > 5ms
		govisual.WithMaxProfileMetrics(500),               // Store up to 500 metrics
	)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("🚀 Server started at http://localhost%s", addr)
	log.Printf("🔍 GoVisual dashboard: http://localhost%s/__viz", addr)
	log.Printf("📊 Performance profiling enabled!")
	log.Println("\n📌 Try these endpoints to see different performance characteristics:")
	log.Println("  - http://localhost:8080/api/fast (< 10ms)")
	log.Println("  - http://localhost:8080/api/slow (100-500ms)")
	log.Println("  - http://localhost:8080/api/cpu-intensive (CPU bound)")
	log.Println("  - http://localhost:8080/api/memory-intensive (High memory usage)")
	log.Println("  - http://localhost:8080/api/database (Multiple DB queries)")
	log.Println("  - http://localhost:8080/api/external (External API calls)")
	log.Println("  - http://localhost:8080/api/complex (Mixed workload)")

	log.Fatal(http.ListenAndServe(addr, handler))
}

func initDatabase() {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	createTables := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		stock INTEGER DEFAULT 0
	);

	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		total_price REAL NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (product_id) REFERENCES products(id)
	);
	`

	if _, err := db.Exec(createTables); err != nil {
		log.Fatal(err)
	}

	// Insert sample data
	insertSampleData()
}

func insertSampleData() {
	// Insert users
	for i := 1; i <= 100; i++ {
		_, err := db.Exec(
			"INSERT INTO users (name, email) VALUES (?, ?)",
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
		)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
		}
	}

	// Insert products
	for i := 1; i <= 50; i++ {
		_, err := db.Exec(
			"INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("Product %d", i),
			fmt.Sprintf("Description for product %d", i),
			float64(rand.Intn(1000))+0.99,
			rand.Intn(100),
		)
		if err != nil {
			log.Printf("Error inserting product: %v", err)
		}
	}
}

// Handler functions with different performance characteristics

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Performance Profiling Example</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; }
			h1 { color: #333; }
			.endpoint { margin: 20px 0; padding: 15px; background: #f5f5f5; border-radius: 5px; }
			.endpoint h3 { margin-top: 0; color: #0066cc; }
			.button { display: inline-block; padding: 10px 20px; background: #0066cc; color: white; text-decoration: none; border-radius: 3px; margin: 5px; }
			.button:hover { background: #0052cc; }
		</style>
	</head>
	<body>
		<h1>🚀 GoVisual Performance Profiling Example</h1>
		<p>Click on the endpoints below to generate requests with different performance characteristics:</p>
		
		<div class="endpoint">
			<h3>Fast Endpoint</h3>
			<p>Returns quickly with minimal processing (&lt; 10ms)</p>
			<a href="/api/fast" class="button">Test Fast</a>
		</div>

		<div class="endpoint">
			<h3>Slow Endpoint</h3>
			<p>Simulates slow processing (100-500ms)</p>
			<a href="/api/slow" class="button">Test Slow</a>
		</div>

		<div class="endpoint">
			<h3>CPU Intensive</h3>
			<p>Performs CPU-bound operations</p>
			<a href="/api/cpu-intensive" class="button">Test CPU</a>
		</div>

		<div class="endpoint">
			<h3>Memory Intensive</h3>
			<p>Allocates significant memory</p>
			<a href="/api/memory-intensive" class="button">Test Memory</a>
		</div>

		<div class="endpoint">
			<h3>Database Operations</h3>
			<p>Performs multiple database queries</p>
			<a href="/api/database" class="button">Test Database</a>
		</div>

		<div class="endpoint">
			<h3>External API</h3>
			<p>Makes external HTTP calls</p>
			<a href="/api/external" class="button">Test External</a>
		</div>

		<div class="endpoint">
			<h3>Complex Workload</h3>
			<p>Mixed operations: CPU, memory, database, and external calls</p>
			<a href="/api/complex" class="button">Test Complex</a>
		</div>

		<hr>
		<p><a href="/__viz" class="button" style="background: #10B981;">Open GoVisual Dashboard</a></p>
	</body>
	</html>
	`
	w.Write([]byte(html))
}

func fastHandler(w http.ResponseWriter, r *http.Request) {
	// Minimal processing
	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now(),
		"message":   "This is a fast endpoint",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate slow processing
	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(delay)

	response := map[string]interface{}{
		"status":    "success",
		"delay_ms":  delay.Milliseconds(),
		"timestamp": time.Now(),
		"message":   "This endpoint intentionally took time",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func cpuIntensiveHandler(w http.ResponseWriter, r *http.Request) {
	// Perform CPU-intensive calculation
	start := time.Now()

	// Calculate prime numbers up to N
	n := 50000
	primes := sieveOfEratosthenes(n)

	duration := time.Since(start)

	response := map[string]interface{}{
		"status":         "success",
		"primes_found":   len(primes),
		"calculation_ms": duration.Milliseconds(),
		"last_10_primes": primes[len(primes)-10:],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func memoryIntensiveHandler(w http.ResponseWriter, r *http.Request) {
	// Allocate significant memory
	size := 10 * 1024 * 1024 // 10 MB
	data := make([]byte, size)

	// Fill with random data
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}

	// Create multiple allocations
	arrays := make([][]int, 100)
	for i := range arrays {
		arrays[i] = make([]int, 10000)
		for j := range arrays[i] {
			arrays[i][j] = rand.Intn(1000)
		}
	}

	response := map[string]interface{}{
		"status":         "success",
		"allocated_mb":   size / 1024 / 1024,
		"arrays_count":   len(arrays),
		"total_elements": len(arrays) * len(arrays[0]),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseHandler(w http.ResponseWriter, r *http.Request) {
	// Query 1: Get users
	var users []User
	start := time.Now()
	rows, err := db.Query("SELECT id, name, email, created_at FROM users LIMIT 10")
	queryDuration := time.Since(start)
	_ = queryDuration // Track for metrics

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u User
			rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
			users = append(users, u)
		}
	}

	// Query 2: Get products
	var products []Product
	start = time.Now()
	rows, err = db.Query("SELECT id, name, description, price, stock FROM products WHERE stock > 0 LIMIT 10")
	queryDuration = time.Since(start)
	_ = queryDuration // Track for metrics

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p Product
			rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock)
			products = append(products, p)
		}
	}

	// Query 3: Count total orders (potentially slow)
	var orderCount int
	start = time.Now()
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&orderCount)
	queryDuration = time.Since(start)
	_ = queryDuration // Track for metrics

	response := map[string]interface{}{
		"status":      "success",
		"users":       users,
		"products":    products,
		"order_count": orderCount,
		"queries_run": 3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func externalAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Make external HTTP calls
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Call 1: Get public IP
	resp1, err := client.Get("https://api.ipify.org?format=json")
	var ip string
	if err == nil {
		defer resp1.Body.Close()
		var result map[string]string
		json.NewDecoder(resp1.Body).Decode(&result)
		ip = result["ip"]
	}

	// Call 2: Get random user
	resp2, err := client.Get("https://randomuser.me/api/")
	var userData interface{}
	if err == nil {
		defer resp2.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&result)
		userData = result
	}

	response := map[string]interface{}{
		"status":      "success",
		"public_ip":   ip,
		"random_user": userData,
		"api_calls":   2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func complexHandler(w http.ResponseWriter, r *http.Request) {
	// Combine multiple operations
	results := make(map[string]interface{})

	// 1. CPU work
	primes := sieveOfEratosthenes(10000)
	results["primes_calculated"] = len(primes)

	// 2. Memory allocation
	data := make([][]byte, 10)
	for i := range data {
		data[i] = make([]byte, 1024*1024) // 1MB each
	}
	results["memory_allocated_mb"] = len(data)

	// 3. Database queries
	var userCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	results["user_count"] = userCount

	// 4. Simulate some delay
	time.Sleep(50 * time.Millisecond)

	// 5. External API call (simulated)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://httpbin.org/delay/1")
	if err == nil {
		resp.Body.Close()
		results["external_api_status"] = resp.StatusCode
	}

	response := map[string]interface{}{
		"status":  "success",
		"results": results,
		"message": "Complex operation completed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function: Sieve of Eratosthenes for prime calculation
func sieveOfEratosthenes(n int) []int {
	if n < 2 {
		return []int{}
	}

	isPrime := make([]bool, n+1)
	for i := 2; i <= n; i++ {
		isPrime[i] = true
	}

	for i := 2; i*i <= n; i++ {
		if isPrime[i] {
			for j := i * i; j <= n; j += i {
				isPrime[j] = false
			}
		}
	}

	primes := []int{}
	for i := 2; i <= n; i++ {
		if isPrime[i] {
			primes = append(primes, i)
		}
	}

	return primes
}
