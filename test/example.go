package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/doganarif/govisual"
)

func main() {
	mux := http.NewServeMux()

	// Add your routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Example handler logic
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Hello, user!"}`))
	})

	// add delay simulation route
	mux.HandleFunc("/dalay/{duration}", func(w http.ResponseWriter, r *http.Request) {
		durationStr := r.PathValue("duration")
		duration, err := strconv.Atoi(durationStr)
		if err != nil || duration < 0 {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Delay completed"}`))
	})

	// add status code simulation route
	mux.HandleFunc("/status/{code}", func(w http.ResponseWriter, r *http.Request) {
		codeStr := r.PathValue("code")
		code, err := strconv.Atoi(codeStr)
		if err != nil || code < 100 || code > 599 {
			http.Error(w, "Invalid status code", http.StatusBadRequest)
			return
		}
		w.WriteHeader(code)
		w.Write([]byte(`{"message": "Status code set"}`))
	})

	// add status code simulation route with delay
	mux.HandleFunc("/status/{code}/delay/{duration}", func(w http.ResponseWriter, r *http.Request) {
		codeStr := r.PathValue("code")
		code, err := strconv.Atoi(codeStr)
		if err != nil || code < 100 || code > 599 {
			http.Error(w, "Invalid status code", http.StatusBadRequest)
			return
		}
		durationStr := r.PathValue("duration")
		duration, err := strconv.Atoi(durationStr)
		if err != nil || duration < 0 {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		w.WriteHeader(code)
		w.Write([]byte(`{"message": "Status code set with delay"}`))
	})

	// Wrap with GoVisual
	handler := govisual.Wrap(
		mux,
		govisual.WithRequestBodyLogging(true),
		govisual.WithResponseBodyLogging(true),
		govisual.WithConsoleLogging(true),
	)

	http.ListenAndServe(":8080", handler)
}
