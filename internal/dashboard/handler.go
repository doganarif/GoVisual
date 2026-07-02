package dashboard

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

//go:embed static/*
var staticFiles embed.FS

// HandlerOptions controls which side-channel endpoints the dashboard exposes.
// The defaults are deliberately restrictive: replay and system-info are
// disabled because they are SSRF / information-disclosure primitives when the
// dashboard is reachable by an attacker.
type HandlerOptions struct {
	// EnableReplay opens POST /api/replay.
	EnableReplay bool
	// ExposeSystemInfo opens GET /api/system-info.
	ExposeSystemInfo bool
	// ExposeEnvVars is the explicit allowlist of env var names the
	// system-info endpoint will surface. Anything not in this set is omitted
	// entirely so an attacker cannot infer existence.
	ExposeEnvVars []string
}

// Handler is the HTTP handler for the dashboard
type Handler struct {
	store       store.Store
	profiler    *profiling.Profiler
	staticFS    fs.FS
	opts        HandlerOptions
	envAllowSet map[string]struct{}
}

// NewHandler creates a new dashboard handler
func NewHandler(store store.Store, profiler *profiling.Profiler, opts HandlerOptions) *Handler {
	staticFS, _ := fs.Sub(staticFiles, "static")

	envSet := make(map[string]struct{}, len(opts.ExposeEnvVars))
	for _, k := range opts.ExposeEnvVars {
		envSet[k] = struct{}{}
	}

	return &Handler{
		store:       store,
		profiler:    profiler,
		staticFS:    staticFS,
		opts:        opts,
		envAllowSet: envSet,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		switch r.URL.Path {
		case "/api/requests":
			h.handleAPIRequests(w, r)
		case "/api/events":
			h.handleSSE(w, r)
		case "/api/clear":
			h.handleClearRequests(w, r)
		case "/api/compare":
			h.handleCompareRequests(w, r)
		case "/api/replay":
			if !h.opts.EnableReplay {
				http.Error(w, "replay disabled", http.StatusNotFound)
				return
			}
			h.handleReplayRequest(w, r)
		case "/api/metrics":
			h.handleMetrics(w, r)
		case "/api/flamegraph":
			h.handleFlameGraph(w, r)
		case "/api/bottlenecks":
			h.handleBottlenecks(w, r)
		case "/api/system-info":
			if !h.opts.ExposeSystemInfo {
				http.Error(w, "system-info disabled", http.StatusNotFound)
				return
			}
			h.handleSystemInfo(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		}
		return
	}

	filePath := r.URL.Path
	if filePath == "/" || filePath == "" {
		filePath = "index.html"
	} else {
		filePath = strings.TrimPrefix(filePath, "/")
	}

	file, err := h.staticFS.Open(filePath)
	if err != nil {
		file, err = h.staticFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		filePath = "index.html"
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch {
	case strings.HasSuffix(filePath, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(filePath, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(filePath, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}

	http.ServeContent(w, r, filePath, stat.ModTime(), file.(io.ReadSeeker))
}

func (h *Handler) handleAPIRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	requests := h.store.GetAll()
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(requests)
}

func (h *Handler) handleClearRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := h.store.Clear(); err != nil {
		http.Error(w, "Error clearing requests", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

// handleSSE streams updates as Server-Sent Events. It sends a full snapshot on
// connect and then publishes only the IDs of the most recent requests on each
// tick — clients diff that against what they already have, so the bandwidth
// scales with churn rather than the entire log.
func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	writeEvent := func(event string, payload interface{}) bool {
		data, err := json.Marshal(payload)
		if err != nil {
			return false
		}
		if event != "" {
			if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
				return false
			}
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// When the store is notifying (the default via Wrap), new entries are
	// pushed as they arrive; the ticker degrades to a heartbeat and a
	// safety-net resync. Subscribe before taking the snapshot so an Add in
	// between is never silently missed — it just triggers one no-op flush.
	var notify <-chan struct{}
	if ns, ok := h.store.(*store.NotifyingStore); ok {
		ch, cancel := ns.Subscribe()
		defer cancel()
		notify = ch
	}

	initial := h.store.GetAll()
	if !writeEvent("snapshot", initial) {
		return
	}
	tick := 15 * time.Second
	if notify == nil {
		tick = 2 * time.Second
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	// Seed lastSeen from the snapshot we just sent. Without this, the first
	// tick would re-emit every entry as an "append" event because the
	// "lastSeen == \"\"" branch treats the whole latest list as new.
	lastSeen := ""
	if len(initial) > 0 {
		lastSeen = initial[0].ID
	}

	// flushNew announces entries added since lastSeen. heartbeat controls
	// whether an idle pass emits a keep-alive comment.
	flushNew := func(heartbeat bool) bool {
		latest := h.store.GetLatest(50)
		// Find any entries newer than what we last announced. The store
		// returns newest-first, so we slice everything before lastSeen.
		found := lastSeen == ""
		cutoff := len(latest)
		for i, l := range latest {
			if l.ID == lastSeen {
				cutoff = i
				found = true
				break
			}
		}
		if lastSeen != "" && !found {
			// lastSeen is no longer in the store — the user cleared the
			// log (or it rolled out of the cap). Resync the client with a
			// fresh snapshot so it discards the stale entries.
			if !writeEvent("snapshot", latest) {
				return false
			}
			if len(latest) > 0 {
				lastSeen = latest[0].ID
			} else {
				lastSeen = ""
			}
			return true
		}
		if cutoff == 0 {
			if !heartbeat {
				return true
			}
			// Heartbeat keeps proxies from closing idle connections.
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return false
			}
			flusher.Flush()
			return true
		}
		fresh := latest[:cutoff]
		if !writeEvent("append", fresh) {
			return false
		}
		lastSeen = fresh[0].ID
		return true
	}

	for {
		select {
		case <-notify:
			if !flushNew(false) {
				return
			}
		case <-ticker.C:
			if !flushNew(true) {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

// maxCompareIDs caps how many request IDs a single /api/compare call may
// supply. Without this, a caller could send thousands of IDs and force one
// store.Get per ID — which on SQL backends is a separate round-trip and a
// cheap amplification primitive.
const maxCompareIDs = 32

func (h *Handler) handleCompareRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ids := r.URL.Query()["id"]
	if len(ids) < 2 {
		http.Error(w, "At least two request IDs are required", http.StatusBadRequest)
		return
	}
	if len(ids) > maxCompareIDs {
		http.Error(w, fmt.Sprintf("too many ids (max %d)", maxCompareIDs), http.StatusBadRequest)
		return
	}

	// Look each id up directly rather than walking the entire log.
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	compareRequests := make([]interface{}, 0, len(ids))
	for id := range idSet {
		if req, ok := h.store.Get(id); ok {
			compareRequests = append(compareRequests, req)
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(compareRequests)
}

// handleReplayRequest replays a captured HTTP request against an arbitrary
// destination. This is a powerful primitive and is therefore opt-in via
// HandlerOptions.EnableReplay. Even when enabled, we deny:
//   - non-http(s) schemes (gopher://, file://, ftp://, etc.)
//   - hostnames that resolve to loopback / link-local / private / multicast IPs
//
// to mitigate SSRF against cloud metadata services or internal networks. Any
// caller that needs to point replay at internal hosts is expected to manage
// network policy themselves; we will not undo the deny-by-default.
func (h *Handler) handleReplayRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var replayRequest struct {
		RequestID string            `json:"requestId"`
		URL       string            `json:"url"`
		Method    string            `json:"method"`
		Headers   map[string]string `json:"headers"`
		Body      string            `json:"body"`
	}
	if err := decoder.Decode(&replayRequest); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateReplayTarget(replayRequest.URL); err != nil {
		http.Error(w, "Replay target rejected: "+err.Error(), http.StatusForbidden)
		return
	}

	// Block redirects — a 30x to a private IP would defeat the pre-flight
	// check. Use a custom DialContext that re-validates the resolved IP at
	// dial time so DNS-rebinding can't slip past the pre-flight check (the
	// pre-flight resolves and validates, but DefaultTransport would otherwise
	// resolve again from the OS cache moments later).
	transport := &http.Transport{
		DialContext: safeDialContext,
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, replayRequest.Method, replayRequest.URL, strings.NewReader(replayRequest.Body))
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	for key, value := range replayRequest.Headers {
		req.Header.Add(key, value)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime).Milliseconds()
	if err != nil {
		http.Error(w, "Error executing request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Cap the captured response body so a hostile target can't OOM us.
	const maxReplayBody = 1 << 20 // 1 MiB
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayBody))
	if err != nil {
		http.Error(w, "Error reading response body: "+err.Error(), http.StatusBadGateway)
		return
	}

	headers := make(map[string][]string, len(resp.Header))
	for k, v := range resp.Header {
		headers[k] = v
	}

	replayResponse := struct {
		StatusCode      int                 `json:"statusCode"`
		Headers         map[string][]string `json:"headers"`
		Body            string              `json:"body"`
		Duration        int64               `json:"duration"`
		OriginalRequest string              `json:"originalRequest"`
	}{
		StatusCode:      resp.StatusCode,
		Headers:         headers,
		Body:            string(respBody),
		Duration:        duration,
		OriginalRequest: replayRequest.RequestID,
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(replayResponse); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// validateReplayTarget rejects replay URLs that point at unsafe schemes or at
// IPs the caller almost certainly did not mean to expose: loopback, link-local,
// multicast, private ranges, and (critically on cloud) the IMDS address.
func validateReplayTarget(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme %q not allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("missing host")
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup failed: %w", err)
	}
	for _, ip := range ips {
		if isInternalIP(ip) {
			return fmt.Errorf("target resolves to non-public address %s", ip)
		}
	}
	return nil
}

// isInternalIP reports whether ip is one we should refuse to dial from a
// replay endpoint. It normalizes IPv4-mapped IPv6 addresses (::ffff:a.b.c.d)
// to their IPv4 form so an attacker cannot bypass the check by encoding a
// private IPv4 address as IPv6.
func isInternalIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() || ip.IsPrivate() {
		return true
	}
	// AWS / GCP / Azure IMDS endpoint.
	if ip.Equal(net.IPv4(169, 254, 169, 254)) {
		return true
	}
	return false
}

// safeDialContext resolves the host and rejects the dial if any resolved
// address is private/loopback/IMDS. Crucially, the same resolution result is
// used for the actual connection — this closes the DNS-rebinding TOCTOU
// window between a pre-flight LookupIP and the transport's own resolution.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses for %s", host)
	}
	for _, ip := range ips {
		if isInternalIP(ip.IP) {
			return nil, fmt.Errorf("dial rejected: %s resolves to non-public address %s", host, ip.IP)
		}
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	// Dial the first resolved address directly so the kernel does not perform
	// a second lookup that could race with the validation above.
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestID := r.URL.Query().Get("id")
	if requestID == "" {
		if h.profiler == nil {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{"error":"Profiling is not enabled"}`))
			return
		}
		metrics := h.profiler.GetAllMetrics()
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.Encode(metrics)
		return
	}

	if h.profiler == nil {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(`{"error":"Profiling is not enabled"}`))
		return
	}

	var payload *store.PerformanceMetrics
	if m, found := h.profiler.GetMetrics(requestID); found {
		payload = m.Model()
	} else {
		reqLog, found := h.store.Get(requestID)
		if !found || reqLog.PerformanceMetrics == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Metrics not found"}`))
			return
		}
		payload = reqLog.PerformanceMetrics
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(payload)
}

func (h *Handler) handleFlameGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestID := r.URL.Query().Get("id")
	if requestID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Request ID is required"}`))
		return
	}

	var cpuProfile []byte
	found := false
	if h.profiler != nil {
		if m, ok := h.profiler.GetMetrics(requestID); ok {
			cpuProfile = m.CPUProfile
			found = true
		}
	}

	if !found {
		reqLog, ok := h.store.Get(requestID)
		if !ok || reqLog.PerformanceMetrics == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Metrics not found for this request"}`))
			return
		}
		cpuProfile = reqLog.PerformanceMetrics.CPUProfile
	}

	if len(cpuProfile) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"No CPU profile data available"}`))
		return
	}

	flameGraph, err := profiling.GenerateFlameGraph(cpuProfile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":"Failed to generate flame graph: %v"}`, err)))
		return
	}

	d3Data := flameGraph.ConvertToD3Format()
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(d3Data)
}

// maxBottleneckScan bounds how many recent requests handleBottlenecks scans.
// The store contract caps capacity already (see MaxRequests), but on shared
// SQL/Mongo backends the table can contain entries from other producers, and
// GetAll on those backends has no LIMIT. Use GetLatest to keep the work
// O(constant) regardless of table size.
const maxBottleneckScan = 500

func (h *Handler) handleBottlenecks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	allRequests := h.store.GetLatest(maxBottleneckScan)

	type BottleneckSummary struct {
		RequestID   string             `json:"request_id"`
		Path        string             `json:"path"`
		Method      string             `json:"method"`
		Duration    int64              `json:"duration"`
		Bottlenecks []store.Bottleneck `json:"bottlenecks"`
	}

	var summaries []BottleneckSummary
	for _, req := range allRequests {
		if req.PerformanceMetrics != nil && len(req.PerformanceMetrics.Bottlenecks) > 0 {
			summaries = append(summaries, BottleneckSummary{
				RequestID:   req.ID,
				Path:        req.Path,
				Method:      req.Method,
				Duration:    req.Duration,
				Bottlenecks: req.PerformanceMetrics.Bottlenecks,
			})
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(summaries)
}

// handleSystemInfo exposes coarse runtime info plus an *explicit allowlist* of
// env vars. The previous implementation used a denylist of substrings ("KEY",
// "SECRET", ...), which is fragile: anything not on the list — DATABASE_URL,
// SLACK_WEBHOOK_URL, JWT_SIGNING_KEY before the bot learns the new abbreviation
// — leaks. Allowlists fail closed.
func (h *Handler) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hostname, _ := os.Hostname()

	envVars := make(map[string]string, len(h.envAllowSet))
	for name := range h.envAllowSet {
		if v, ok := os.LookupEnv(name); ok {
			envVars[name] = v
		}
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := map[string]interface{}{
		"goVersion":   runtime.Version(),
		"goos":        runtime.GOOS,
		"goarch":      runtime.GOARCH,
		"hostname":    hostname,
		"cpuCores":    runtime.NumCPU(),
		"memoryUsed":  memStats.Alloc / 1024 / 1024,
		"memoryTotal": memStats.Sys / 1024 / 1024,
		"envVars":     envVars,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(systemInfo)
}
