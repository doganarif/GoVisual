package govisual

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/doganarif/govisual/internal/dashboard"
	"github.com/doganarif/govisual/internal/middleware"
	"github.com/doganarif/govisual/internal/profiling"
	"github.com/doganarif/govisual/internal/store"
	"github.com/doganarif/govisual/internal/telemetry"
)

// Wrap wraps an http.Handler with the govisual request visualization middleware
// and mounts the dashboard at config.DashboardPath. Pass options to customize
// behavior. To trigger graceful shutdown of storage and telemetry resources,
// pass WithShutdownContext — govisual will release its resources when that
// context is cancelled. Govisual deliberately does NOT register a signal
// handler; that is the host application's job.
func Wrap(handler http.Handler, opts ...Option) http.Handler {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	var requestStore store.Store
	storeConfig := &store.StorageConfig{
		Type:             config.StorageType,
		Capacity:         config.MaxRequests,
		ConnectionString: config.ConnectionString,
		TableName:        config.TableName,
		TTL:              config.RedisTTL,
		ExistingDB:       config.ExistingDB,
	}
	rs, err := store.NewStore(storeConfig)
	if err != nil {
		log.Printf("govisual: failed to create configured storage backend: %v. Falling back to in-memory storage.", err)
		requestStore = store.NewInMemoryStore(config.MaxRequests)
	} else {
		requestStore = rs
	}

	var profiler *profiling.Profiler
	if config.EnableProfiling {
		profiler = profiling.NewProfiler(config.MaxProfileMetrics)
		profiler.SetEnabled(true)
		profiler.SetProfileType(config.ProfileType)
		profiler.SetThreshold(config.ProfileThreshold)
		log.Printf("govisual: performance profiling enabled (threshold=%v)", config.ProfileThreshold)
	}

	var wrapped http.Handler
	if profiler != nil {
		wrapped = middleware.WrapWithProfilingAndLimits(
			handler, requestStore,
			config.LogRequestBody, config.LogResponseBody,
			config, profiler, config.effectiveMaxBody(),
		)
	} else {
		wrapped = middleware.WrapWithLimits(
			handler, requestStore,
			config.LogRequestBody, config.LogResponseBody,
			config, config.effectiveMaxBody(),
		)
	}

	var otelShutdown func(context.Context) error
	if config.EnableOpenTelemetry {
		ctx := context.Background()
		otelConfig := telemetry.Config{
			ServiceName:    config.ServiceName,
			ServiceVersion: config.ServiceVersion,
			Endpoint:       config.OTelEndpoint,
			Insecure:       config.OTelInsecure,
			Exporter:       config.OTelExporter,
		}
		shutdown, err := telemetry.InitTracer(ctx, otelConfig)
		if err != nil {
			log.Printf("govisual: failed to initialize OpenTelemetry: %v", err)
		} else {
			log.Printf("govisual: OpenTelemetry initialized (service=%s endpoint=%s)", config.ServiceName, config.OTelEndpoint)
			otelShutdown = shutdown
			wrapped = middleware.NewOTelMiddleware(wrapped, config.ServiceName, config.ServiceVersion)
		}
	}

	if config.ShutdownContext != nil {
		// NOTE: this goroutine waits on ctx.Done() and is retained for the
		// process lifetime if the context is never cancelled. Callers passing a
		// non-cancellable context (e.g. context.Background()) should be aware
		// of this — in tests, prefer t.Context() or a cancellable context.
		go func(ctx context.Context, st store.Store, shutdown func(context.Context) error) {
			<-ctx.Done()
			log.Printf("govisual: shutdown context cancelled, releasing resources")
			if shutdown != nil {
				// Give OTel a real deadline to flush spans, independent of
				// the parent context (which is already cancelled).
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := shutdown(shutdownCtx); err != nil {
					log.Printf("govisual: error shutting down OpenTelemetry: %v", err)
				}
				cancel()
			}
			if err := st.Close(); err != nil {
				log.Printf("govisual: error closing storage: %v", err)
			}
		}(config.ShutdownContext, requestStore, otelShutdown)
	}

	dashHandler := dashboard.NewHandler(requestStore, profiler, dashboard.HandlerOptions{
		EnableReplay:     config.EnableReplay,
		ExposeSystemInfo: config.ExposeSystemInfo,
		ExposeEnvVars:    config.ExposeEnvVars,
	})

	guardedDash := guardDashboard(dashHandler, config)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == config.DashboardPath {
			// The dashboard uses relative URLs, which only resolve under
			// the trailing-slash form of the mount path.
			http.Redirect(w, r, config.DashboardPath+"/", http.StatusMovedPermanently)
			return
		}
		if strings.HasPrefix(r.URL.Path, config.DashboardPath+"/") {
			http.StripPrefix(config.DashboardPath, guardedDash).ServeHTTP(w, r)
			return
		}
		wrapped.ServeHTTP(w, r)
	})
}

// guardDashboard wraps the dashboard handler with localhost-only and
// authentication checks per the configuration.
func guardDashboard(h http.Handler, config *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.LocalhostOnly && !isLoopback(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if config.DashboardAuth != nil && !config.DashboardAuth(r) {
			// Surface a Basic challenge so browsers prompt the user; harmless
			// when a custom auth scheme is in use.
			w.Header().Set("WWW-Authenticate", `Basic realm="govisual"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// isLoopback reports whether the request's remote address is a loopback IP.
func isLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}

// effectiveMaxBody resolves the configured MaxBodyBytes against the package
// default. 0 means "use default"; negative means "no cap".
func (c *Config) effectiveMaxBody() int {
	if c.MaxBodyBytes == 0 {
		return middleware.DefaultMaxBodyBytes
	}
	return c.MaxBodyBytes
}
