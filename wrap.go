package govisual

import (
	"github.com/doganarif/govisual/internal/dashboard"
	"github.com/doganarif/govisual/internal/middleware"
	"github.com/doganarif/govisual/internal/store"
	"net/http"
	"strings"
)

// Wrap wraps an http.Handler with request visualization middleware
func Wrap(handler http.Handler, opts ...Option) http.Handler {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create store
	requestStore := store.NewInMemoryStore(config.MaxRequests)

	// Create middleware wrapper
	wrapped := middleware.Wrap(handler, requestStore, config.LogRequestBody, config.LogResponseBody, config)

	// Create dashboard handler
	dashHandler := dashboard.NewHandler(requestStore)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, config.DashboardPath) {
			dashPath := strings.TrimPrefix(r.URL.Path, config.DashboardPath)
			if dashPath == "" {
				http.Redirect(w, r, config.DashboardPath+"/", http.StatusFound)
				return
			}
			http.StripPrefix(config.DashboardPath, dashHandler).ServeHTTP(w, r)
			return
		}

		// Otherwise, serve the application
		wrapped.ServeHTTP(w, r)
	})
}
