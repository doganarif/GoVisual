package govisual

import (
	"path/filepath"
	"strings"
)

type Config struct {
	MaxRequests int

	DashboardPath string

	LogRequestBody bool

	LogResponseBody bool

	IgnorePaths []string
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
		c.DashboardPath = path
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

// WithIgnorePaths sets the path patterns to ignore
func WithIgnorePaths(patterns ...string) Option {
	return func(c *Config) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
func (c *Config) ShouldIgnorePath(path string) bool {
	// First check if it's the dashboard path which should always be ignored to prevent recursive logging
	if path == c.DashboardPath || strings.HasPrefix(path, c.DashboardPath+"/") {
		return true
	}

	// Then check against provided ignore patterns
	for _, pattern := range c.IgnorePaths {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Special handling for path groups with trailing slash
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			// If pattern ends with /, check if path starts with pattern
			if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
				return true
			}
		}
	}

	return false
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		MaxRequests:     100,
		DashboardPath:   "/__viz",
		LogRequestBody:  false,
		LogResponseBody: false,
		IgnorePaths:     []string{},
	}
}
