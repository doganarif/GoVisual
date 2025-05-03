package main

import (
	"fmt"
	"github.com/doganarif/govisual"
	"log"
	"net/http"
	"time"
)

func main() {
	// Create a simple mux
	mux := http.NewServeMux()

	// Add some test endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the test server! Visit /__viz to see the dashboard.")
	})

	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	mux.HandleFunc("/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello, JSON!"}`)
	})

	mux.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintf(w, "This was a slow request!")
	})

	mux.HandleFunc("/api/very-slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 2000ms - clearly visible
		fmt.Fprintf(w, "This was a very slow request! (2000ms)")
	})

	mux.HandleFunc("/api/medium-slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // 1000ms
		fmt.Fprintf(w, "This was a medium slow request! (1000ms)")
	})

	mux.HandleFunc("/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "This is an error response!")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "# Sample metrics data\nrequests_total 42\n")
	})

	// Wrap with govisual
	handler := govisual.Wrap(mux,
		govisual.WithMaxRequests(200),
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		// Ignore health checks, metrics endpoints, and all static assets
		govisual.WithIgnorePaths(
			"/health",
			"/metrics",
			"/static/*",      // Wildcard for all static files
			"/api/internal/", // Ignore all paths under this directory
			"/favicon.ico",
		),
	)

	// Start the server
	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println("Visit http://localhost:8080/__viz to see the dashboard")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
