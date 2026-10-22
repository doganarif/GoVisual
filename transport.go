package govisual

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/doganarif/govisual/v2/internal/profiling"
)

// WrapTransport instruments an http.RoundTripper so outbound calls made
// while handling a request show up on that request's profile:
//
//	client := &http.Client{Transport: govisual.WrapTransport(nil)}
//	resp, err := client.Do(req.WithContext(r.Context()))
//
// Attribution goes through the request context, so outbound requests must
// carry it, and profiling must be enabled via WithProfiling(true). A nil rt
// wraps http.DefaultTransport.
// Body capture is opt-in via WithTransportBodyLogging and is bounded by
// WithTransportMaxBodyBytes. Bodies are captured as they are read, without
// eagerly consuming request streams or delaying response headers.
func WrapTransport(rt http.RoundTripper, opts ...TransportOption) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	t := &vizTransport{base: rt, maxBodyBytes: 1 << 20}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// TransportOption configures outbound HTTP capture in WrapTransport.
type TransportOption func(*vizTransport)

// WithTransportBodyLogging enables request and response body capture for
// outbound HTTP calls. Disabled by default. Captured bodies may contain
// sensitive application data; enable only when appropriate for your app.
func WithTransportBodyLogging(enabled bool) TransportOption {
	return func(t *vizTransport) { t.logBodies = enabled }
}

// WithTransportMaxBodyBytes limits each captured outbound body to max bytes.
// Zero or negative values use the default 1 MiB limit. The actual HTTP body
// is always passed through in full. Bodies exceeding the limit are marked
// as truncated in the dashboard.
func WithTransportMaxBodyBytes(max int) TransportOption {
	return func(t *vizTransport) {
		t.maxBodyBytes = 1 << 20
		if max > 0 {
			t.maxBodyBytes = max
		}
	}
}

type vizTransport struct {
	base         http.RoundTripper
	logBodies    bool
	maxBodyBytes int
}

func (t *vizTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	call := profiling.BeginHTTPCall(req.Context(), req.Method, req.URL.String())
	var requestBody *transportBody
	if t.logBodies && call != nil {
		call.SetBody(true, "", false)
		if req.Body != nil && req.Body != http.NoBody {
			requestBody = &transportBody{ReadCloser: req.Body, max: t.maxBodyBytes,
				record: func(body string, truncated bool) { call.SetBody(true, body, truncated) }}
			req = req.Clone(req.Context())
			req.Body = requestBody
		}
	}
	resp, err := t.base.RoundTrip(req)
	duration := time.Since(start)
	if requestBody != nil {
		requestBody.snapshot()
	}

	status := 0
	size := int64(0)
	if resp != nil {
		status = resp.StatusCode
		if resp.ContentLength > 0 {
			size = resp.ContentLength
		}
	}
	if call != nil {
		call.Complete(duration, status, size)
		// Preserve the bidirectional body returned for protocol upgrades.
		if t.logBodies && resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
			call.SetBody(false, "", false)
			if resp.Body != nil && resp.Body != http.NoBody {
				resp.Body = &transportBody{ReadCloser: resp.Body, max: t.maxBodyBytes,
					record: func(body string, truncated bool) { call.SetBody(false, body, truncated) }}
			}
		}
	}
	return resp, err
}

// transportBody observes only bytes consumed by the transport/caller. Close
// still reaches the original body, including while a Read is blocked.
type transportBody struct {
	io.ReadCloser
	mu        sync.Mutex
	buffer    bytes.Buffer
	max       int
	truncated bool
	record    func(string, bool)
}

func (b *transportBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.mu.Lock()
	remaining := b.max - b.buffer.Len()
	if n > remaining {
		b.truncated = true
	}
	if remaining > 0 {
		b.buffer.Write(p[:min(n, remaining)])
	}
	b.mu.Unlock()
	if err != nil {
		b.snapshot()
	}
	return n, err
}

func (b *transportBody) Close() error {
	err := b.ReadCloser.Close()
	b.snapshot()
	return err
}

func (b *transportBody) snapshot() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.record(b.buffer.String(), b.truncated)
}
