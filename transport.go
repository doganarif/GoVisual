package govisual

import (
	"net/http"
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
func WrapTransport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &vizTransport{base: rt}
}

type vizTransport struct {
	base http.RoundTripper
}

func (t *vizTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	duration := time.Since(start)

	status := 0
	size := int64(0)
	if resp != nil {
		status = resp.StatusCode
		if resp.ContentLength > 0 {
			size = resp.ContentLength
		}
	}
	profiling.RecordHTTP(req.Context(), req.Method, req.URL.String(), duration, status, size)
	return resp, err
}
