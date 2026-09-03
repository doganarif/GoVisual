package dashboard

import (
	"context"
	"embed"
	"encoding/json"
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
	// ReplayBaseURL pins replay traffic to an application URL. When empty, a
	// localhost-only handler may use its validated request scheme and Host.
	ReplayBaseURL string
	// AllowLocalhostReplay permits the request Host fallback, but only when it
	// names localhost or a literal loopback IP. Remote dashboards must instead
	// configure ReplayBaseURL explicitly.
	AllowLocalhostReplay bool
	// ExposeSystemInfo opens GET /api/system-info.
	ExposeSystemInfo bool
	// ExposeEnvVars is the explicit allowlist of env var names the
	// system-info endpoint will surface. Anything not in this set is omitted
	// entirely so an attacker cannot infer existence.
	ExposeEnvVars []string
	// ActivityLog, if set, powers /api/agent-activity.
	ActivityLog *store.ActivityLog
}

const captureTruncationMarker = "...[truncated by govisual]"

const (
	maxReplayRequestBody = 1 << 20
	maxReplayPayload     = 8 << 20
	maxReplayHeaders     = 100
	maxReplayHeaderBytes = 64 << 10
)

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
		case "/api/agent-activity":
			h.handleAgentActivity(w, r)
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

// handleReplayRequest loads a captured request by ID, applies shape-only
// overrides, and replays it against the application origin. The client cannot
// supply or alter the destination authority.
func (h *Handler) handleReplayRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReplayPayload)
	decoder := json.NewDecoder(r.Body)
	var replayRequest struct {
		RequestID string `json:"requestId"`
		// URL accepts the v2.0.0 dashboard payload for compatibility. Its
		// authority is ignored; only its path and query can be replayed.
		URL     *string           `json:"url,omitempty"`
		Method  *string           `json:"method,omitempty"`
		Path    *string           `json:"path,omitempty"`
		Headers map[string]string `json:"headers"`
		Body    *string           `json:"body,omitempty"`
	}
	if err := decoder.Decode(&replayRequest); err != nil {
		http.Error(w, "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if replayRequest.RequestID == "" {
		http.Error(w, "Request ID is required", http.StatusBadRequest)
		return
	}
	original, ok := h.store.Get(replayRequest.RequestID)
	if !ok {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	method := original.Method
	if replayRequest.Method != nil && *replayRequest.Method != "" {
		method = *replayRequest.Method
	}
	body := original.RequestBody
	if strings.HasSuffix(original.RequestBody, captureTruncationMarker) &&
		(replayRequest.Body == nil || *replayRequest.Body == original.RequestBody) {
		http.Error(w, "Captured request body is truncated; provide the complete body before replaying", http.StatusBadRequest)
		return
	}
	if replayRequest.Body != nil {
		body = *replayRequest.Body
	}
	if len(body) > maxReplayRequestBody {
		http.Error(w, "Replay request body exceeds 1 MiB", http.StatusRequestEntityTooLarge)
		return
	}
	pathOverride := replayRequest.Path
	if pathOverride == nil && replayRequest.URL != nil {
		legacyURL, err := url.Parse(*replayRequest.URL)
		if err != nil || legacyURL == nil || (legacyURL.Scheme != "http" && legacyURL.Scheme != "https") || legacyURL.Host == "" || legacyURL.User != nil || legacyURL.Fragment != "" {
			http.Error(w, "Invalid replay target: invalid legacy URL", http.StatusBadRequest)
			return
		}
		legacyPath := legacyURL.EscapedPath()
		if legacyPath == "" {
			legacyPath = "/"
		}
		if legacyURL.RawQuery != "" {
			legacyPath += "?" + legacyURL.RawQuery
		}
		pathOverride = &legacyPath
	}
	target, err := h.replayTarget(r, original, pathOverride)
	if err != nil {
		http.Error(w, "Invalid replay target: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Preserve captured Accept-Encoding semantics and exact response bytes;
	// automatic decompression would otherwise silently change the replay.
	transport := &http.Transport{Proxy: nil, DisableCompression: true}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(body))
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if replayRequest.Headers == nil {
		// An omitted header map is the legacy "replay unchanged" shape.
		copyReplayHeaders(req.Header, original.RequestHeaders, nil)
	} else {
		// The dashboard sends the complete edited map, so an absent key means
		// the user removed that captured header.
		copyReplayHeaders(req.Header, nil, replayRequest.Headers)
	}
	req.Header.Set("X-Govisual-Replay", "1")
	if err := validateReplayHeaders(req.Header); err != nil {
		http.Error(w, "Invalid replay headers: "+err.Error(), http.StatusBadRequest)
		return
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
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxReplayBody+1))
	if err != nil {
		http.Error(w, "Error reading response body: "+err.Error(), http.StatusBadGateway)
		return
	}
	bodyTruncated := len(respBody) > maxReplayBody
	if bodyTruncated {
		respBody = respBody[:maxReplayBody]
	}

	headerLog := &store.RequestLog{}
	headerLog.SetResponseHeaders(resp.Header)

	replayResponse := struct {
		StatusCode      int         `json:"statusCode"`
		Headers         http.Header `json:"headers"`
		Body            string      `json:"body"`
		Duration        int64       `json:"duration"`
		OriginalRequest string      `json:"originalRequest"`
		BodyTruncated   bool        `json:"bodyTruncated,omitempty"`
	}{
		StatusCode:      resp.StatusCode,
		Headers:         headerLog.ResponseHeaders,
		Body:            string(respBody),
		Duration:        duration,
		OriginalRequest: replayRequest.RequestID,
		BodyTruncated:   bodyTruncated,
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(replayResponse); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) replayTarget(r *http.Request, original *store.RequestLog, override *string) (string, error) {
	baseRaw := h.opts.ReplayBaseURL
	if baseRaw == "" {
		if !h.opts.AllowLocalhostReplay || !isLocalReplayHost(r.Host) {
			return "", fmt.Errorf("replay base URL is required unless the dashboard is localhost-only")
		}
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseRaw = scheme + "://" + r.Host
	}
	base, err := url.Parse(baseRaw)
	if err != nil || base == nil {
		return "", fmt.Errorf("invalid replay base URL")
	}
	base.Scheme = strings.ToLower(base.Scheme)
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil || base.Opaque != "" || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" {
		return "", fmt.Errorf("invalid replay base URL")
	}

	requestPath := original.Path
	requestRawPath := original.RawPath
	rawQuery := original.Query
	if override != nil {
		relative, err := url.Parse(*override)
		if err != nil || relative.IsAbs() || relative.Host != "" || relative.Fragment != "" || !strings.HasPrefix(relative.Path, "/") {
			return "", fmt.Errorf("path must be an absolute path without a host")
		}
		requestPath = relative.Path
		requestRawPath = relative.RawPath
		rawQuery = relative.RawQuery
	}
	if requestPath == "" || !strings.HasPrefix(requestPath, "/") {
		return "", fmt.Errorf("captured request has an invalid path")
	}

	target := *base
	target.Path = strings.TrimSuffix(base.Path, "/") + requestPath
	if requestRawPath != "" {
		target.RawPath = strings.TrimSuffix(base.EscapedPath(), "/") + requestRawPath
	} else {
		target.RawPath = ""
	}
	target.RawQuery = rawQuery
	return target.String(), nil
}

var fixedHopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Proxy-Connection":    {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
	"Host":                {},
	"Content-Length":      {},
}

func copyReplayHeaders(dst http.Header, captured http.Header, overrides map[string]string) {
	blocked := make(map[string]struct{}, len(fixedHopByHopHeaders))
	for name := range fixedHopByHopHeaders {
		blocked[name] = struct{}{}
	}
	for key, values := range captured {
		if http.CanonicalHeaderKey(key) == "Connection" {
			blockConnectionTokens(blocked, values...)
		}
	}
	for key, value := range overrides {
		if http.CanonicalHeaderKey(key) == "Connection" {
			blockConnectionTokens(blocked, value)
		}
	}

	for key, values := range captured {
		canonical := http.CanonicalHeaderKey(key)
		if _, skip := blocked[canonical]; skip || canonical == "" {
			continue
		}
		for _, value := range values {
			if value != "[redacted by govisual]" {
				dst.Add(canonical, value)
			}
		}
	}
	for key, value := range overrides {
		canonical := http.CanonicalHeaderKey(key)
		if _, skip := blocked[canonical]; skip || canonical == "" {
			continue
		}
		if value == "[redacted by govisual]" {
			continue
		}
		dst.Set(canonical, value)
	}
}

func isLocalReplayHost(authority string) bool {
	u, err := url.Parse("http://" + authority)
	if err != nil || u.User != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4
		}
		return ip.IsLoopback()
	}
	return host == "localhost" || strings.HasSuffix(host, ".localhost")
}

func blockConnectionTokens(blocked map[string]struct{}, values ...string) {
	for _, value := range values {
		for _, token := range strings.Split(value, ",") {
			if canonical := http.CanonicalHeaderKey(strings.TrimSpace(token)); canonical != "" {
				blocked[canonical] = struct{}{}
			}
		}
	}
}

func validateReplayHeaders(headers http.Header) error {
	if len(headers) > maxReplayHeaders {
		return fmt.Errorf("too many headers (maximum %d)", maxReplayHeaders)
	}
	remaining := maxReplayHeaderBytes
	for key, values := range headers {
		if len(key) > remaining {
			return fmt.Errorf("headers exceed %d bytes", maxReplayHeaderBytes)
		}
		remaining -= len(key)
		for _, value := range values {
			if len(value) > remaining {
				return fmt.Errorf("headers exceed %d bytes", maxReplayHeaderBytes)
			}
			remaining -= len(value)
		}
	}
	return nil
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

// handleAgentActivity returns the most recent coding-agent tool calls. Empty
// when no ActivityLog is configured, so the UI can render "no activity yet"
// without needing a separate probe.
func (h *Handler) handleAgentActivity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var entries []store.ActivityEntry
	if h.opts.ActivityLog != nil {
		entries = h.opts.ActivityLog.List(200)
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(entries)
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
