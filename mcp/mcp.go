// Package mcp serves govisual's captured traffic to AI coding agents over
// the Model Context Protocol. Mount the handler next to (not inside) the
// wrapped application so agent traffic isn't captured as requests:
//
//	st := store.NewMemory(200)
//	app := govisual.Wrap(mux, govisual.WithStore(st))
//
//	root := http.NewServeMux()
//	root.Handle("/mcp", gvmcp.Handler(st, gvmcp.WithBaseURL("http://localhost:8080")))
//	root.Handle("/", app)
//
// Then point a client at it: claude mcp add govisual --transport http http://localhost:8080/mcp
package mcp

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type config struct {
	baseURL     string
	allowRemote bool
	token       string
	activity    *store.ActivityLog
}

// Option configures the MCP handler.
type Option func(*config)

var errInvalidAuthority = errors.New("invalid host")

// WithBaseURL sets the URL replays are sent to. Replay always targets the
// wrapped application — an agent can change method, path, headers, and body,
// but never the destination host, so the endpoint is not an SSRF primitive.
// It is required for replay, diff, and generated curl commands.
func WithBaseURL(u string) Option {
	return func(c *config) { c.baseURL = u }
}

// WithAllowRemote lets non-loopback addresses use the MCP endpoint. Off by
// default for the same reason the dashboard is loopback-only. Always pair it
// with WithToken or equivalent authentication in an outer handler.
func WithAllowRemote() Option {
	return func(c *config) { c.allowRemote = true }
}

// WithToken requires a Bearer token on every MCP request.
func WithToken(token string) Option {
	return func(c *config) { c.token = token }
}

// WithActivityLog records every tool call into the given log so the
// govisual dashboard can surface them on its Agents tab. Pass the same
// *store.ActivityLog to govisual.Wrap via govisual.WithActivityLog.
func WithActivityLog(a *store.ActivityLog) Option {
	return func(c *config) { c.activity = a }
}

// Handler returns the MCP endpoint for a govisual store. Use the same store
// instance you passed to govisual.Wrap via WithStore.
func Handler(st store.Store, opts ...Option) http.Handler {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	srv := newServer(st, cfg)
	mcpHandler := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return srv }, nil)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestScheme := "http"
		if r.TLS != nil {
			requestScheme = "https"
		}
		if _, _, err := canonicalAuthority(r.Host, requestScheme); err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !cfg.allowRemote && (!isLoopback(r) || !isLocalAuthority(r.Host)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !hasSameOrigin(r, cfg.allowRemote) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if cfg.token != "" && r.Header.Get("Authorization") != "Bearer "+cfg.token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mcpHandler.ServeHTTP(w, r)
	})
}

func newServer(st store.Store, cfg *config) *sdk.Server {
	srv := sdk.NewServer(&sdk.Implementation{Name: "govisual", Version: "2.0.1"}, nil)
	registerReadTools(srv, st, cfg)
	registerReplayTools(srv, st, cfg)
	registerActionTools(srv, st, cfg)
	return srv
}

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
		// A remote MCP endpoint may sit behind a TLS-terminating proxy. Match
		// its public Host and port without trusting forwarded headers.
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
		return "", "", errInvalidAuthority
	}
	u, err := url.Parse(scheme + "://" + authority)
	if err != nil || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", errInvalidAuthority
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" {
		return "", "", errInvalidAuthority
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
			return "", "", errInvalidAuthority
		}
		port = strconv.Itoa(n)
	}
	if ip := net.ParseIP(host); ip != nil {
		host = ip.String()
	}
	return host, port, nil
}
