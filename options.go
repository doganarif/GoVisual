package govisual

import (
	"context"
	"crypto/subtle"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

// DashboardAuth authorizes a request to the dashboard. Return true to allow,
// false to deny (govisual sends an HTTP 401). Implementations should be
// constant-time when comparing secrets.
type DashboardAuth func(r *http.Request) bool

type Config struct {
	MaxRequests int

	DashboardPath string

	LogRequestBody bool

	LogResponseBody bool

	// MaxBodyBytes caps the captured request and response body size.
	// 0 (default) means use middleware.DefaultMaxBodyBytes (1 MiB).
	// Set to -1 to disable the cap entirely (NOT recommended).
	MaxBodyBytes int

	IgnorePaths []string

	// OpenTelemetry configuration
	EnableOpenTelemetry bool

	ServiceName string

	ServiceVersion string

	OTelEndpoint string

	OTelInsecure bool

	OTelExporter string

	// Store is the storage backend for captured requests. Nil means an
	// in-memory store bounded by MaxRequests. Database-backed stores live in
	// their own modules under store/ (postgres, redis, sqlite, mongodb).
	Store store.Store

	// Performance Profiling configuration
	EnableProfiling bool

	ProfileType ProfileType

	ProfileThreshold time.Duration

	MaxProfileMetrics int

	// Dashboard security ----------------------------------------------------

	// DashboardAuth, if set, must approve every request to the dashboard.
	// If nil, the dashboard is fully open — only safe for local dev.
	DashboardAuth DashboardAuth

	// LocalhostOnly, when true, rejects dashboard requests whose remote address
	// is not a loopback IP. This is the safest default for "I'm just debugging
	// locally" — even with the rest of the server bound to 0.0.0.0.
	LocalhostOnly bool

	// EnableReplay enables the POST /__viz/api/replay endpoint, which lets the
	// dashboard fire arbitrary HTTP requests from the server. Disabled by
	// default because it is a powerful SSRF primitive if the dashboard is
	// reachable by an attacker.
	EnableReplay bool

	// ExposeSystemInfo controls whether the GET /__viz/api/system-info endpoint
	// is enabled. Disabled by default; enabling exposes runtime info (hostname,
	// Go version, memory stats).
	ExposeSystemInfo bool

	// ExposeEnvVars is an explicit allowlist of environment variable names that
	// the system-info endpoint may surface. Anything not in this set is omitted
	// entirely (NOT redacted) so an attacker cannot infer the existence of a
	// sensitive name.
	ExposeEnvVars []string

	// ShutdownContext, if set, triggers graceful shutdown of govisual-owned
	// resources (storage backends, OpenTelemetry tracer provider) when the
	// context is cancelled. This replaces the prior behavior of registering a
	// global signal handler that called os.Exit — a library has no business
	// killing the host process.
	ShutdownContext context.Context
}

// Option is a function that modifies the configuration
type Option func(*Config)

// WithMaxRequests sets the maximum number of requests to store
func WithMaxRequests(max int) Option {
	return func(c *Config) {
		c.MaxRequests = max
	}
}

// WithDashboardPath sets the path to access the dashboard
func WithDashboardPath(path string) Option {
	return func(c *Config) {
		c.DashboardPath = strings.TrimSuffix(path, "/")
	}
}

// WithRequestBodyLogging enables or disables request body logging
func WithRequestBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogRequestBody = enabled
	}
}

// WithResponseBodyLogging enables or disables response body logging
func WithResponseBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogResponseBody = enabled
	}
}

// WithMaxBodyBytes caps the captured request and response body size.
// Values:
//   - 0: use the package default (1 MiB)
//   - >0: explicit cap in bytes
//   - <0: disable cap (unbounded — be careful with large downloads)
func WithMaxBodyBytes(n int) Option {
	return func(c *Config) {
		c.MaxBodyBytes = n
	}
}

// WithIgnorePaths sets the path patterns to ignore
func WithIgnorePaths(patterns ...string) Option {
	return func(c *Config) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// WithOpenTelemetry enables or disables OpenTelemetry instrumentation
func WithOpenTelemetry(enabled bool) Option {
	return func(c *Config) {
		c.EnableOpenTelemetry = enabled
	}
}

// WithServiceName sets the service name for OpenTelemetry
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion sets the service version for OpenTelemetry
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithOTelEndpoint sets the OTLP endpoint for exporting telemetry data
func WithOTelEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.OTelEndpoint = endpoint
	}
}

// WithOTelInsecure sets whether to use an insecure connection for OTLP
func WithOTelInsecure(insecure bool) Option {
	return func(c *Config) {
		c.OTelInsecure = insecure
	}
}

// WithOTelExporter sets the type of exporter to use.
// Valid values: "otlp" (default), "stdout" (for debugging), "noop" (for benchmarking)
func WithOTelExporter(exporterType string) Option {
	return func(c *Config) {
		c.OTelExporter = exporterType
	}
}

// WithStore sets the storage backend for captured requests. Construct one
// from a storage module, e.g. postgres.New(...) from
// github.com/doganarif/govisual/v2/store/postgres. Without this option an
// in-memory store bounded by WithMaxRequests is used.
func WithStore(s store.Store) Option {
	return func(c *Config) {
		c.Store = s
	}
}

// ShouldIgnorePath checks if a path should be ignored based on the configured patterns.
func (c *Config) ShouldIgnorePath(path string) bool {
	// The dashboard itself must always be ignored, otherwise opening it
	// would recursively log every poll.
	if path == c.DashboardPath || strings.HasPrefix(path, c.DashboardPath+"/") {
		return true
	}

	for _, pattern := range c.IgnorePaths {
		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}
		// Trailing-slash patterns are treated as "prefix match".
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			if strings.HasPrefix(path, pattern) {
				return true
			}
		}
	}
	return false
}

// WithProfiling enables or disables performance profiling
func WithProfiling(enabled bool) Option {
	return func(c *Config) {
		c.EnableProfiling = enabled
	}
}

// WithProfileType sets the types of profiling to perform
func WithProfileType(profileType ProfileType) Option {
	return func(c *Config) {
		c.ProfileType = profileType
	}
}

// WithProfileThreshold sets the minimum duration to trigger profiling
func WithProfileThreshold(threshold time.Duration) Option {
	return func(c *Config) {
		c.ProfileThreshold = threshold
	}
}

// WithMaxProfileMetrics sets the maximum number of profile metrics to store
func WithMaxProfileMetrics(max int) Option {
	return func(c *Config) {
		c.MaxProfileMetrics = max
	}
}

// WithDashboardAuth installs a custom authentication function for the dashboard.
// The function runs on every dashboard request and must return true to allow access.
func WithDashboardAuth(fn DashboardAuth) Option {
	return func(c *Config) {
		c.DashboardAuth = fn
	}
}

// WithBasicAuth protects the dashboard with HTTP Basic Auth using a constant-time
// comparison. Both username and password are required.
func WithBasicAuth(username, password string) Option {
	expectedUser := []byte(username)
	expectedPass := []byte(password)
	return func(c *Config) {
		c.DashboardAuth = func(r *http.Request) bool {
			user, pass, ok := r.BasicAuth()
			if !ok {
				return false
			}
			userOK := subtle.ConstantTimeCompare([]byte(user), expectedUser) == 1
			passOK := subtle.ConstantTimeCompare([]byte(pass), expectedPass) == 1
			return userOK && passOK
		}
	}
}

// WithLocalhostOnly restricts the dashboard to requests originating from a
// loopback address. Combine with WithDashboardAuth/WithBasicAuth for defense
// in depth.
func WithLocalhostOnly() Option {
	return func(c *Config) {
		c.LocalhostOnly = true
	}
}

// WithReplayEnabled enables the dashboard's /api/replay endpoint. Disabled by
// default because the endpoint, if reachable, lets a caller make the server
// perform arbitrary outbound HTTP requests (an SSRF primitive). Only enable
// behind authentication and/or localhost-only access.
func WithReplayEnabled(enabled bool) Option {
	return func(c *Config) {
		c.EnableReplay = enabled
	}
}

// WithSystemInfo enables the dashboard's /api/system-info endpoint and
// optionally sets the allowlist of environment variable names to expose.
// Pass no names to enable the endpoint but expose nothing (memory/runtime
// info only).
func WithSystemInfo(envAllowlist ...string) Option {
	return func(c *Config) {
		c.ExposeSystemInfo = true
		c.ExposeEnvVars = append(c.ExposeEnvVars, envAllowlist...)
	}
}

// WithShutdownContext wires govisual's internal cleanup (storage backends,
// OpenTelemetry shutdown) to a caller-provided context. When the context is
// cancelled, govisual releases its resources. Replaces the prior behavior of
// installing a global signal handler that called os.Exit.
//
// Note: govisual spawns one goroutine that blocks on ctx.Done() for the
// lifetime of the wrapped handler. If you never cancel the context (for
// example, by passing context.Background()), that goroutine is retained for
// the process lifetime — harmless in long-running services, but tests should
// pass a cancellable context (e.g. t.Context()) to avoid leaks across cases.
func WithShutdownContext(ctx context.Context) Option {
	return func(c *Config) {
		c.ShutdownContext = ctx
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		MaxRequests:         100,
		DashboardPath:       "/__viz",
		LogRequestBody:      false,
		LogResponseBody:     false,
		MaxBodyBytes:        0, // 0 => use middleware.DefaultMaxBodyBytes
		IgnorePaths:         []string{},
		EnableOpenTelemetry: false,
		ServiceName:         "govisual",
		ServiceVersion:      "dev",
		OTelEndpoint:        "localhost:4317",
		OTelInsecure:        true,
		OTelExporter:        "otlp",
		EnableProfiling:     false,
		ProfileType:         profiling.ProfileAll,
		ProfileThreshold:    10 * time.Millisecond,
		MaxProfileMetrics:   1000,
		EnableReplay:        false,
		ExposeSystemInfo:    false,
	}
}
