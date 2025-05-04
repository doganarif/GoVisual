package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/doganarif/govisual"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.Parse()

	// Create a simple HTTP server
	mux := http.NewServeMux()

	// Add example routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body>
			<h1>GoVisual Basic Example</h1>
			<p>Visit <a href="/__viz">/__viz</a> to access the request visualizer</p>
			<p>API Endpoints:</p>
			<ul>
				<li><a href="/api/hello">/api/hello</a> - Simple JSON response</li>
				<li><a href="/api/slow">/api/slow</a> - Slow response (500ms)</li>
				<li><a href="/api/error">/api/error</a> - Error response</li>
			</ul>
		</body></html>`)
	})

	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, World!",
		})
	})

	mux.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":  "This was slow",
			"duration": "500ms",
		})
	})

	mux.HandleFunc("/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal Server Error",
		})
	})

	// Wrap with GoVisual
	handler := govisual.Wrap(
		mux,
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
	)

	// Start the server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server started at http://localhost%s", addr)
	log.Printf("Visit http://localhost%s/__viz to see the dashboard", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
