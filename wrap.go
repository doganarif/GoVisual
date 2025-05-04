package govisual

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/doganarif/govisual/internal/dashboard"
	"github.com/doganarif/govisual/internal/middleware"
	"github.com/doganarif/govisual/internal/store"
	"github.com/doganarif/govisual/internal/telemetry"
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

	// Initialize OpenTelemetry if enabled
	var shutdown func(context.Context) error
	if config.EnableOpenTelemetry {
		ctx := context.Background()
		var err error
		shutdown, err = telemetry.InitTracer(ctx, config.ServiceName, config.ServiceVersion, config.OTelEndpoint)
		if err != nil {
			log.Printf("Failed to initialize OpenTelemetry: %v", err)
		} else {
			log.Printf("OpenTelemetry initialized with service name: %s, endpoint: %s", config.ServiceName, config.OTelEndpoint)

			// Set up graceful shutdown for OpenTelemetry
			go func() {
				signals := make(chan os.Signal, 1)
				signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
				<-signals
				if err := shutdown(context.Background()); err != nil {
					log.Printf("Error shutting down OpenTelemetry: %v", err)
				}
			}()

			// Wrap with OpenTelemetry middleware
			wrapped = middleware.NewOTelMiddleware(wrapped, config.ServiceName, config.ServiceVersion)
		}
	}

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
