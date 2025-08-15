package govisual

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/doganarif/govisual/internal/dashboard"
	"github.com/doganarif/govisual/internal/middleware"
	"github.com/doganarif/govisual/internal/store"
	"github.com/doganarif/govisual/internal/telemetry"
)

var (
	// Global signal handler to ensure we only have one
	signalOnce    sync.Once
	shutdownFuncs []func(context.Context) error
	shutdownMutex sync.Mutex
)

// addShutdownFunc adds a shutdown function to be called on signal
func addShutdownFunc(fn func(context.Context) error) {
	if fn == nil {
		log.Println("Warning: Attempted to register nil shutdown function, ignoring")
		return
	}
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()
	shutdownFuncs = append(shutdownFuncs, fn)
}

// setupSignalHandler sets up a single signal handler for all cleanup operations
func setupSignalHandler() {
	signalOnce.Do(func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			sig := <-signals
			log.Printf("Received shutdown signal (%v), cleaning up...", sig)

			ctx := context.Background()
			shutdownMutex.Lock()
			funcs := make([]func(context.Context) error, len(shutdownFuncs))
			copy(funcs, shutdownFuncs)
			shutdownMutex.Unlock()

			// Execute all shutdown functions
			for _, fn := range funcs {
				if err := fn(ctx); err != nil {
					log.Printf("Error during shutdown: %v", err)
				}
			}

			log.Println("Cleanup completed, exiting...")

			// Stop listening for more signals and exit
			signal.Stop(signals)
			os.Exit(0)
		}()
	})
}

// WrapWebSocket wraps a WebSocket handler with request visualization middleware
func WrapWebSocket(handler http.HandlerFunc, opts ...Option) http.Handler {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create store based on configuration
	var requestStore store.Store
	var err error

	storeConfig := &store.StorageConfig{
		Type:             config.StorageType,
		Capacity:         config.MaxRequests,
		ConnectionString: config.ConnectionString,
		TableName:        config.TableName,
		TTL:              config.RedisTTL,
		ExistingDB:       config.ExistingDB,
	}

	requestStore, err = store.NewStore(storeConfig)
	if err != nil {
		log.Printf("Failed to create configured storage backend: %v. Falling back to in-memory storage.", err)
		requestStore = store.NewInMemoryStore(config.MaxRequests)
	}

	// Add store cleanup to shutdown functions
	addShutdownFunc(func(ctx context.Context) error {
		if err := requestStore.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
			return err
		}
		return nil
	})

	// Create WebSocket wrapper options
	wsOpts := &middleware.WebSocketWrapperOptions{
		LogMessageBody: config.LogRequestBody,
	}

	// Create WebSocket wrapper
	wrapped := middleware.WrapWebSocket(handler, requestStore, config, wsOpts)

	// Initialize OpenTelemetry if enabled
	if config.EnableOpenTelemetry {
		ctx := context.Background()
		shutdown, err := telemetry.InitTracer(ctx, config.ServiceName, config.ServiceVersion, config.OTelEndpoint)
		if err != nil {
			log.Printf("Failed to initialize OpenTelemetry: %v", err)
		} else {
			log.Printf("OpenTelemetry initialized with service name: %s, endpoint: %s", config.ServiceName, config.OTelEndpoint)

			// Add OpenTelemetry shutdown to shutdown functions
			addShutdownFunc(shutdown)

			// Wrap with OpenTelemetry middleware
			wrapped = middleware.NewOTelMiddleware(wrapped, config.ServiceName, config.ServiceVersion)
		}
	}

	// Set up the single signal handler
	setupSignalHandler()

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

		// Otherwise, serve the WebSocket application
		wrapped.ServeHTTP(w, r)
	})
}

// Wrap wraps an http.Handler with request visualization middleware
func Wrap(handler http.Handler, opts ...Option) http.Handler {
	// Apply options to default config
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create store based on configuration
	var requestStore store.Store
	var err error

	storeConfig := &store.StorageConfig{
		Type:             config.StorageType,
		Capacity:         config.MaxRequests,
		ConnectionString: config.ConnectionString,
		TableName:        config.TableName,
		TTL:              config.RedisTTL,
		ExistingDB:       config.ExistingDB,
	}

	requestStore, err = store.NewStore(storeConfig)
	if err != nil {
		log.Printf("Failed to create configured storage backend: %v. Falling back to in-memory storage.", err)
		requestStore = store.NewInMemoryStore(config.MaxRequests)
	}

	// Add store cleanup to shutdown functions
	addShutdownFunc(func(ctx context.Context) error {
		if err := requestStore.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
			return err
		}
		return nil
	})

	// Create middleware wrapper
	wrapped := middleware.Wrap(handler, requestStore, config.LogRequestBody, config.LogResponseBody, config)

	// Initialize OpenTelemetry if enabled
	if config.EnableOpenTelemetry {
		ctx := context.Background()
		shutdown, err := telemetry.InitTracer(ctx, config.ServiceName, config.ServiceVersion, config.OTelEndpoint)
		if err != nil {
			log.Printf("Failed to initialize OpenTelemetry: %v", err)
		} else {
			log.Printf("OpenTelemetry initialized with service name: %s, endpoint: %s", config.ServiceName, config.OTelEndpoint)

			// Add OpenTelemetry shutdown to shutdown functions
			addShutdownFunc(shutdown)

			// Wrap with OpenTelemetry middleware
			wrapped = middleware.NewOTelMiddleware(wrapped, config.ServiceName, config.ServiceVersion)
		}
	}

	// Set up the single signal handler
	setupSignalHandler()

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
