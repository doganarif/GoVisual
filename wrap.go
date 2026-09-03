package govisual

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/doganarif/govisual/v2/internal/dashboard"
	"github.com/doganarif/govisual/v2/internal/middleware"
	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

// Wrap wraps an http.Handler with the govisual request visualization middleware
// and mounts the dashboard at config.DashboardPath. Pass options to customize
// behavior. To trigger graceful shutdown of the storage backend, pass
// WithShutdownContext — govisual will release its resources when that context
// is cancelled. Govisual deliberately does NOT register a signal handler;
// that is the host application's job.
func Wrap(handler http.Handler, opts ...Option) http.Handler {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	var requestStore store.Store = config.Store
	if requestStore == nil {
		requestStore = store.NewMemory(config.MaxRequests)
	}
	// Notification lets the dashboard push live updates instead of polling.
	// Callers that already wrapped their store (e.g. to share it with the
	// MCP handler's await_request) are left alone.
	if _, ok := requestStore.(*store.NotifyingStore); !ok {
		requestStore = store.WithNotify(requestStore)
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
		wrapped = middleware.WrapWithProfilingLimitsAndErrorHandler(
			handler, requestStore,
			config.LogRequestBody, config.LogResponseBody,
			config, profiler, config.effectiveMaxBody(), config.SampleRate,
			func(err error) { config.handleError(err) },
		)
	} else {
		wrapped = middleware.WrapWithLimitsAndErrorHandler(
			handler, requestStore,
			config.LogRequestBody, config.LogResponseBody,
			config, config.effectiveMaxBody(), config.SampleRate,
			func(err error) { config.handleError(err) },
		)
	}

	if config.ShutdownContext != nil {
		// NOTE: this goroutine waits on ctx.Done() and is retained for the
		// process lifetime if the context is never cancelled. Callers passing a
		// non-cancellable context (e.g. context.Background()) should be aware
		// of this — in tests, prefer t.Context() or a cancellable context.
		go func(ctx context.Context, st store.Store) {
			<-ctx.Done()
			log.Printf("govisual: shutdown context cancelled, releasing resources")
			if err := st.Close(); err != nil {
				log.Printf("govisual: error closing storage: %v", err)
			}
		}(config.ShutdownContext, requestStore)
	}

	dashHandler := dashboard.NewHandler(requestStore, profiler, dashboard.HandlerOptions{
		EnableReplay:         config.EnableReplay,
		ReplayBaseURL:        config.ReplayBaseURL,
		AllowLocalhostReplay: config.LocalhostOnly,
		ExposeSystemInfo:     config.ExposeSystemInfo,
		ExposeEnvVars:        config.ExposeEnvVars,
		ActivityLog:          config.ActivityLog,
	})

	dashboardRoutes := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == config.DashboardPath {
			// The dashboard uses relative URLs, which only resolve under the
			// trailing-slash form of the mount path.
			http.Redirect(w, r, config.DashboardPath+"/", http.StatusMovedPermanently)
			return
		}
		http.StripPrefix(config.DashboardPath, dashHandler).ServeHTTP(w, r)
	})
	guardedDash := guardDashboard(dashboardRoutes, config)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == config.DashboardPath || strings.HasPrefix(r.URL.Path, config.DashboardPath+"/") {
			guardedDash.ServeHTTP(w, r)
			return
		}
		wrapped.ServeHTTP(w, r)
	})
}

// guardDashboard wraps the dashboard handler with localhost-only and
// authentication checks per the configuration.
func guardDashboard(h http.Handler, config *Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestScheme := "http"
		if r.TLS != nil {
			requestScheme = "https"
		}
		if _, _, err := canonicalAuthority(r.Host, requestScheme); err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if config.LocalhostOnly && (!isLoopback(r) || !isLocalAuthority(r.Host)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !hasSameOrigin(r, !config.LocalhostOnly) {
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

func (c *Config) handleError(err error) {
	if c.ErrorHandler != nil {
		c.ErrorHandler(err)
		return
	}
	log.Printf("%v", err)
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

func isLocalAuthority(authority string) bool {
	host, _, err := canonicalAuthority(authority, "http")
	if err != nil {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4
		}
		return ip.IsLoopback()
	}
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

// hasSameOrigin rejects browser requests whose Origin does not exactly match
// the request scheme and Host. Non-browser clients commonly omit Origin and
// are governed by the loopback/authentication checks instead.
func hasSameOrigin(r *http.Request, allowTLSProxy bool) bool {
	origins := r.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}
	if len(origins) != 1 {
		return false
	}

	u, err := url.Parse(origins[0])
	if err != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return false
	}
	originScheme := strings.ToLower(u.Scheme)
	if originScheme != "http" && originScheme != "https" {
		return false
	}
	requestScheme := "http"
	if r.TLS != nil {
		requestScheme = "https"
	}
	if originScheme != requestScheme {
		// WithAllowRemote is commonly used behind a TLS-terminating reverse
		// proxy, where Go sees HTTP even though the browser origin is HTTPS.
		// Accept that one transition without trusting spoofable forwarded
		// headers; the public Host and effective port must still match below.
		if !allowTLSProxy || requestScheme != "http" || originScheme != "https" {
			return false
		}
		requestScheme = originScheme
	}

	originHost, originPort, err := canonicalAuthority(u.Host, originScheme)
	if err != nil {
		return false
	}
	requestHost, requestPort, err := canonicalAuthority(r.Host, requestScheme)
	if err != nil {
		return false
	}
	return originHost == requestHost && originPort == requestPort
}

func canonicalAuthority(authority, scheme string) (string, string, error) {
	authority = strings.TrimSpace(authority)
	if authority == "" {
		return "", "", fmt.Errorf("empty host")
	}
	u, err := url.Parse(scheme + "://" + authority)
	if err != nil || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", fmt.Errorf("invalid host")
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" {
		return "", "", fmt.Errorf("empty host")
	}
	port := u.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	} else {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return "", "", fmt.Errorf("invalid port")
		}
		port = strconv.Itoa(n)
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	return host, port, nil
}

// effectiveMaxBody resolves the configured MaxBodyBytes against the package
// default. 0 means "use default"; negative means "no cap".
func (c *Config) effectiveMaxBody() int {
	if c.MaxBodyBytes == 0 {
		return middleware.DefaultMaxBodyBytes
	}
	return c.MaxBodyBytes
}
