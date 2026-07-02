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
	"net"
	"net/http"

	"github.com/doganarif/govisual/v2/store"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type config struct {
	baseURL     string
	allowRemote bool
	token       string
}

// Option configures the MCP handler.
type Option func(*config)

// WithBaseURL sets the URL replays are sent to. Replay always targets the
// wrapped application — an agent can change method, path, headers, and body,
// but never the destination host, so the endpoint is not an SSRF primitive.
// Defaults to http://<captured Host> of the request being replayed.
func WithBaseURL(u string) Option {
	return func(c *config) { c.baseURL = u }
}

// WithAllowRemote lets non-loopback addresses use the MCP endpoint. Off by
// default for the same reason the dashboard is loopback-only.
func WithAllowRemote() Option {
	return func(c *config) { c.allowRemote = true }
}

// WithToken requires a Bearer token on every MCP request. Recommended
// whenever WithAllowRemote is used.
func WithToken(token string) Option {
	return func(c *config) { c.token = token }
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
		if !cfg.allowRemote && !isLoopback(r) {
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
	srv := sdk.NewServer(&sdk.Implementation{Name: "govisual", Version: "2.0.0"}, nil)
	registerReadTools(srv, st)
	registerReplayTools(srv, st, cfg)
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
