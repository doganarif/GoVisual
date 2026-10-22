package govisual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func transportProfile(t *testing.T) (context.Context, *profiling.Profiler) {
	t.Helper()
	p := profiling.NewProfiler(10)
	p.SetProfileType(profiling.ProfileNone)
	p.SetThreshold(0)
	return p.StartProfiling(context.Background(), t.Name()), p
}

func TestTransportBodyCaptureEndToEnd(t *testing.T) {
	const requestBody = `{"name":"gopher"}`
	const responseBody = `{"id":42,"name":"gopher"}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != requestBody {
			t.Errorf("upstream body = %q, %v", body, err)
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, responseBody)
	}))
	defer upstream.Close()

	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			st := store.NewMemory(10)
			client := &http.Client{Transport: WrapTransport(nil, WithTransportBodyLogging(enabled))}
			h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, upstream.URL, strings.NewReader(requestBody))
				resp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				io.Copy(w, resp.Body)
			}), WithStore(st), WithProfiling(true), WithProfileType(profiling.ProfileNone), WithProfileThreshold(0))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/users", nil))
			if rec.Body.String() != responseBody {
				t.Fatalf("caller body = %q", rec.Body.String())
			}
			logs := st.GetAll()
			if len(logs) != 1 || len(logs[0].PerformanceMetrics.HTTPCalls) != 1 {
				t.Fatalf("expected one persisted outbound call: %+v", logs)
			}
			call := logs[0].PerformanceMetrics.HTTPCalls[0]
			if call.Status != http.StatusCreated || call.Method != http.MethodPost {
				t.Fatalf("metadata = %+v", call)
			}
			if enabled {
				if call.RequestBody == nil || *call.RequestBody != requestBody || call.ResponseBody == nil || *call.ResponseBody != responseBody {
					t.Fatalf("persisted bodies = %+v", call)
				}
			} else if call.RequestBody != nil || call.ResponseBody != nil {
				t.Fatal("body capture should be disabled by default")
			}
			// The dashboard API must expose the same optional body fields.
			req := httptest.NewRequest(http.MethodGet, "/__viz/api/requests", nil)
			req.Host, req.RemoteAddr = "localhost", "127.0.0.1:1234"
			apiResponse := httptest.NewRecorder()
			h.ServeHTTP(apiResponse, req)
			var decoded []store.RequestLog
			if err := json.Unmarshal(apiResponse.Body.Bytes(), &decoded); err != nil || apiResponse.Code != 200 {
				t.Fatalf("dashboard API: %d %s, %v", apiResponse.Code, apiResponse.Body, err)
			}
			got := decoded[0].PerformanceMetrics.HTTPCalls[0]
			if enabled && (got.ResponseBody == nil || *got.ResponseBody != responseBody) {
				t.Fatalf("API lost captured body: %+v", got)
			}
		})
	}
}

func TestTransportBodyLimitAndEmptyBodies(t *testing.T) {
	for _, payload := range []string{"", "1234", "123456789"} {
		t.Run(fmt.Sprint(len(payload)), func(t *testing.T) {
			ctx, p := transportProfile(t)
			base := transportFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				req.Body.Close()
				if err != nil || string(body) != payload {
					t.Fatalf("upstream body = %q, %v", body, err)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload))}, nil
			})
			req, _ := http.NewRequestWithContext(ctx, "POST", "http://example.test", strings.NewReader(payload))
			resp, err := WrapTransport(base, WithTransportBodyLogging(true), WithTransportMaxBodyBytes(4)).RoundTrip(req)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil || string(body) != payload {
				t.Fatalf("caller body = %q, %v", body, err)
			}
			call := p.EndProfiling(ctx).HTTPCalls[0]
			want := payload[:min(len(payload), 4)]
			if call.RequestBody == nil || *call.RequestBody != want || call.ResponseBody == nil || *call.ResponseBody != want {
				t.Fatalf("capture = %+v, want %q", call, want)
			}
			if call.RequestBodyTruncated != (len(payload) > 4) || call.ResponseBodyTruncated != (len(payload) > 4) {
				t.Fatalf("truncation flags = %+v", call)
			}
		})
	}
}

type observedBody struct {
	io.Reader
	reads, closes int
	closeErr      error
}

func (b *observedBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *observedBody) Close() error               { b.closes++; return b.closeErr }

func TestTransportStreamingAndPartialClose(t *testing.T) {
	ctx, p := transportProfile(t)
	closeErr := errors.New("close failed")
	original := &observedBody{Reader: strings.NewReader("streaming response"), closeErr: closeErr}
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.test", nil)
	base := transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: original, ContentLength: -1}, nil
	})
	resp, err := WrapTransport(base, WithTransportBodyLogging(true)).RoundTrip(req)
	if err != nil || original.reads != 0 {
		t.Fatalf("RoundTrip eagerly read body: reads=%d err=%v", original.reads, err)
	}
	before, _ := p.GetMetrics(t.Name())
	if before.HTTPCalls[0].Status != 200 {
		t.Fatal("metadata should be available before reading the response")
	}
	buf := make([]byte, 6)
	io.ReadFull(resp.Body, buf)
	if err := resp.Body.Close(); err != closeErr || original.closes != 1 {
		t.Fatalf("Close did not forward: %v, %d", err, original.closes)
	}
	call := p.EndProfiling(ctx).HTTPCalls[0]
	if call.ResponseBody == nil || *call.ResponseBody != "stream" {
		t.Fatalf("partial body = %+v", call)
	}
	if *before.HTTPCalls[0].ResponseBody != "" {
		t.Fatal("a later body update mutated an earlier snapshot")
	}
}

func TestTransportNoProfileAndDisabledCapture(t *testing.T) {
	for _, profiled := range []bool{false, true} {
		t.Run(fmt.Sprint(profiled), func(t *testing.T) {
			ctx := context.Background()
			if profiled {
				ctx, _ = transportProfile(t)
			}
			body := &observedBody{Reader: strings.NewReader("untouched")}
			req, _ := http.NewRequestWithContext(ctx, "POST", "http://example.test", body)
			base := transportFunc(func(got *http.Request) (*http.Response, error) {
				if got != req || got.Body != body {
					t.Fatal("unneeded request wrapping")
				}
				return &http.Response{StatusCode: 200, Body: body}, nil
			})
			resp, _ := WrapTransport(base, WithTransportBodyLogging(!profiled)).RoundTrip(req)
			if resp.Body != body || body.reads != 0 || body.closes != 0 {
				t.Fatal("unneeded body wrapping or reading")
			}
		})
	}
}

func TestTransportPreservesErrors(t *testing.T) {
	ctx, p := transportProfile(t)
	want := errors.New("connection failed")
	base := transportFunc(func(req *http.Request) (*http.Response, error) {
		io.Copy(io.Discard, req.Body)
		req.Body.Close()
		return nil, want
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", "http://example.test", strings.NewReader("payload"))
	originalBody := req.Body
	resp, err := WrapTransport(base, WithTransportBodyLogging(true)).RoundTrip(req)
	if resp != nil || err != want || req.Body != originalBody {
		t.Fatalf("transport changed result or original request: %v, %v", resp, err)
	}
	call := p.EndProfiling(ctx).HTTPCalls[0]
	if call.Status != 0 || call.RequestBody == nil || *call.RequestBody != "payload" || call.ResponseBody != nil {
		t.Fatalf("failed call = %+v", call)
	}
}

type upgradedBody struct{ io.ReadCloser }

func (b *upgradedBody) Write(p []byte) (int, error) { return len(p), nil }

func TestTransportPreservesProtocolUpgrade(t *testing.T) {
	ctx, p := transportProfile(t)
	body := &upgradedBody{io.NopCloser(strings.NewReader("upgraded stream"))}
	base := transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusSwitchingProtocols, Body: body}, nil
	})
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.test", nil)
	resp, err := WrapTransport(base, WithTransportBodyLogging(true)).RoundTrip(req)
	if err != nil || resp.Body != body {
		t.Fatalf("upgrade body was wrapped: %v, %v", resp, err)
	}
	if _, ok := resp.Body.(io.ReadWriteCloser); !ok {
		t.Fatal("upgrade lost its Write method")
	}
	if p.EndProfiling(ctx).HTTPCalls[0].ResponseBody != nil {
		t.Fatal("upgrade stream must not be captured")
	}
}

func TestTransportConcurrentBodyCapture(t *testing.T) {
	ctx, p := transportProfile(t)
	base := transportFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(body)))}, err
	})
	rt := WrapTransport(base, WithTransportBodyLogging(true))
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "POST", "http://example.test", strings.NewReader(fmt.Sprint(i)))
			resp, err := rt.RoundTrip(req)
			if err != nil {
				t.Error(err)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			// Exercise snapshot/model serialization while other calls finish.
			m, _ := p.GetMetrics(t.Name())
			json.Marshal(m.Model())
		}(i)
	}
	wg.Wait()
	calls := p.EndProfiling(ctx).HTTPCalls
	if len(calls) != 20 {
		t.Fatalf("got %d calls", len(calls))
	}
	for _, call := range calls {
		if call.RequestBody == nil || call.ResponseBody == nil || *call.RequestBody != *call.ResponseBody {
			t.Fatalf("mixed up concurrent calls: %+v", call)
		}
	}
}
