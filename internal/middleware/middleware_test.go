package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/doganarif/govisual/v2/internal/profiling"
	"github.com/doganarif/govisual/v2/store"
)

// mockStore implements store.Store for testing
type mockStore struct {
	logs   []*store.RequestLog
	addErr error
}

func (m *mockStore) Add(log *store.RequestLog) error {
	m.logs = append(m.logs, log)
	return m.addErr
}

func (m *mockStore) Get(id string) (*store.RequestLog, bool) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, true
		}
	}
	return nil, false
}

func (m *mockStore) GetAll() []*store.RequestLog {
	return m.logs
}

func (m *mockStore) Clear() error {
	m.logs = nil
	return nil
}

func (m *mockStore) GetLatest(n int) []*store.RequestLog {
	if n >= len(m.logs) {
		return m.logs
	}
	return m.logs[len(m.logs)-n:]
}

func (m *mockStore) Close() error {
	return nil
}

// mockPathMatcher implements PathMatcher
type mockPathMatcher struct{}

func (m *mockPathMatcher) ShouldIgnorePath(path string) bool {
	return false
}

func TestWrapMiddleware(t *testing.T) {
	store := &mockStore{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("hello world"))
	})

	wrapped := Wrap(handler, store, true, true, &mockPathMatcher{})

	req := httptest.NewRequest("POST", "/test?x=1", strings.NewReader("sample-body"))
	req.Header.Set("X-Test", "test")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if len(store.logs) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(store.logs))
	}
	log := store.logs[0]

	if log.Method != "POST" {
		t.Errorf("expected Method POST, got %s", log.Method)
	}
	if log.Path != "/test" {
		t.Errorf("expected Path /test, got %s", log.Path)
	}
	if log.RequestBody != "sample-body" {
		t.Errorf("expected RequestBody to be 'sample-body', got '%s'", log.RequestBody)
	}
	if log.ResponseBody != "hello world" {
		t.Errorf("expected ResponseBody to be 'hello world', got '%s'", log.ResponseBody)
	}
	if log.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, log.StatusCode)
	}
	if log.Duration < 0 {
		t.Errorf("expected Duration > 0, got %d", log.Duration)
	}
}

func TestStatusReadRacesWithLateWrites(t *testing.T) {
	store := &mockStore{}
	var wg sync.WaitGroup

	// A handler that leaks a goroutine still writing after it returns; the
	// middleware's post-handler reads must not race with it.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.Write([]byte("late"))
			}
		}()
	})

	wrapped := Wrap(handler, store, false, true, &mockPathMatcher{})
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	wg.Wait()
}

func TestRequestBodyCaptureDoesNotTruncateDownstreamBody(t *testing.T) {
	st := &mockStore{}
	original := "0123456789"
	var downstream string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("handler read body: %v", err)
		}
		downstream = string(body)
	})

	wrapped := WrapWithLimits(handler, st, true, false, nil, 4, 1)
	wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(original)))

	if downstream != original {
		t.Fatalf("downstream body = %q, want exact original %q", downstream, original)
	}
	if len(st.logs) != 1 || st.logs[0].RequestBody != "0123"+truncationMarker {
		t.Fatalf("captured body = %q, want capped body with marker", st.logs[0].RequestBody)
	}
}

type trackingBody struct {
	*bytes.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

type failingBody struct {
	data []byte
	done bool
}

func (b *failingBody) Read(p []byte) (int, error) {
	if b.done {
		return 0, io.EOF
	}
	b.done = true
	return copy(p, b.data), io.ErrUnexpectedEOF
}

func (*failingBody) Close() error { return nil }

func TestRequestBodyCapturePreservesReadError(t *testing.T) {
	st := &mockStore{}
	wantBody := "partial"
	var downstreamBody string
	var downstreamErr error
	var reportedErr error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		downstreamBody = string(body)
		downstreamErr = err
	})

	wrapped := WrapWithLimitsAndErrorHandler(handler, st, true, false, nil, 1024, 1, func(err error) {
		reportedErr = err
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = &failingBody{data: []byte(wantBody)}
	wrapped.ServeHTTP(httptest.NewRecorder(), req)

	if downstreamBody != wantBody {
		t.Fatalf("downstream body = %q, want %q", downstreamBody, wantBody)
	}
	if !errors.Is(downstreamErr, io.ErrUnexpectedEOF) {
		t.Fatalf("downstream error = %v, want %v", downstreamErr, io.ErrUnexpectedEOF)
	}
	if !errors.Is(reportedErr, io.ErrUnexpectedEOF) {
		t.Fatalf("reported capture error = %v, want %v", reportedErr, io.ErrUnexpectedEOF)
	}
	if len(st.logs) != 1 || st.logs[0].RequestBody != "" {
		t.Fatalf("failed capture should not be stored as a complete body: %+v", st.logs)
	}
}

func TestCaptureRequestBodyReturnsErrorWithFinalReplayedBytes(t *testing.T) {
	want := []byte("partial")
	_, restored, captureErr := captureRequestBody(&failingBody{data: want}, 1024)
	if !errors.Is(captureErr, io.ErrUnexpectedEOF) {
		t.Fatalf("capture error = %v, want %v", captureErr, io.ErrUnexpectedEOF)
	}
	buf := make([]byte, len(want))
	n, err := restored.Read(buf)
	if n != len(want) || !bytes.Equal(buf[:n], want) || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("replayed Read = %d, %q, %v; want %d, %q, %v", n, buf[:n], err, len(want), want, io.ErrUnexpectedEOF)
	}
}

func TestRequestBodyRemainsOpenUntilHandlerClosesIt(t *testing.T) {
	st := &mockStore{}
	body := &trackingBody{Reader: bytes.NewReader([]byte("body"))}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body.closed {
			t.Fatal("middleware closed the request body before the handler ran")
		}
		if err := r.Body.Close(); err != nil {
			t.Fatalf("close body: %v", err)
		}
	})

	wrapped := WrapWithLimits(handler, st, true, false, nil, 2, 1)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = body
	wrapped.ServeHTTP(httptest.NewRecorder(), req)
	if !body.closed {
		t.Fatal("closing reconstructed body did not close original body")
	}
}

func TestResponseHeadersAreCapturedAndRedacted(t *testing.T) {
	st := &mockStore{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Visible", "yes")
		w.Header().Set("Set-Cookie", "session=secret")
		w.Header().Set("Authorization", "Bearer secret")
		w.WriteHeader(http.StatusCreated)
		// Headers changed after WriteHeader are not part of the committed response.
		w.Header().Set("X-Visible", "too-late")
	})

	Wrap(handler, st, false, false, nil).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if len(st.logs) != 1 {
		t.Fatalf("got %d logs, want 1", len(st.logs))
	}
	headers := st.logs[0].ResponseHeaders
	if got := headers.Get("X-Visible"); got != "yes" {
		t.Fatalf("captured X-Visible = %q, want committed value", got)
	}
	for _, name := range []string{"Set-Cookie", "Authorization"} {
		if got := headers.Get(name); got != "[redacted by govisual]" {
			t.Fatalf("captured %s = %q, want redaction", name, got)
		}
	}
}

type shortResponseWriter struct {
	header http.Header
	body   bytes.Buffer
}

func (w *shortResponseWriter) Header() http.Header { return w.header }
func (w *shortResponseWriter) WriteHeader(int)     {}
func (w *shortResponseWriter) Write(p []byte) (int, error) {
	if len(p) > 2 {
		p = p[:2]
	}
	return w.body.Write(p)
}

type headerRecordingWriter struct {
	header http.Header
	codes  []int
}

func (w *headerRecordingWriter) Header() http.Header { return w.header }
func (w *headerRecordingWriter) WriteHeader(code int) {
	w.codes = append(w.codes, code)
	if code >= 200 {
		w.header.Set("Content-Encoding", "gzip")
	}
}
func (w *headerRecordingWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestResponseWriterCapturesOnlyWrittenBytesAndUnwraps(t *testing.T) {
	underlying := &shortResponseWriter{header: make(http.Header)}
	w := &responseWriter{ResponseWriter: underlying, buffer: &bytes.Buffer{}, maxBody: 1024}
	n, err := w.Write([]byte("abcdef"))
	if err != nil || n != 2 {
		t.Fatalf("Write = %d, %v; want 2, nil", n, err)
	}
	if got := w.body(); got != "ab" {
		t.Fatalf("captured body = %q, want only written bytes", got)
	}
	if w.Unwrap() != underlying {
		t.Fatal("Unwrap did not return underlying writer")
	}
}

func TestResponseWriterPreservesOptionalInterfaceSet(t *testing.T) {
	underlying := &shortResponseWriter{header: make(http.Header)}
	wrapped := wrapResponseWriter(&responseWriter{ResponseWriter: underlying, statusCode: http.StatusOK})
	if _, ok := wrapped.(http.Flusher); ok {
		t.Fatal("wrapper invented http.Flusher")
	}
	if _, ok := wrapped.(http.Hijacker); ok {
		t.Fatal("wrapper invented http.Hijacker")
	}
	if _, ok := wrapped.(http.Pusher); ok {
		t.Fatal("wrapper invented http.Pusher")
	}
	if err := http.NewResponseController(wrapped).Flush(); !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("ResponseController.Flush error = %v, want ErrNotSupported", err)
	}

	recorderWrapped := wrapResponseWriter(&responseWriter{ResponseWriter: httptest.NewRecorder(), statusCode: http.StatusOK})
	if _, ok := recorderWrapped.(http.Flusher); !ok {
		t.Fatal("wrapper dropped supported http.Flusher")
	}
}

func TestResponseWriterCapturesFinalStatusAndCommittedHeaders(t *testing.T) {
	underlying := &headerRecordingWriter{header: make(http.Header)}
	w := &responseWriter{ResponseWriter: underlying, statusCode: http.StatusOK}
	w.WriteHeader(http.StatusEarlyHints)
	w.WriteHeader(http.StatusCreated)

	if got := w.status(); got != http.StatusCreated {
		t.Fatalf("captured status = %d, want %d", got, http.StatusCreated)
	}
	if got := w.headers().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("captured committed header = %q, want gzip", got)
	}
	if len(underlying.codes) != 2 || underlying.codes[0] != http.StatusEarlyHints || underlying.codes[1] != http.StatusCreated {
		t.Fatalf("forwarded statuses = %v", underlying.codes)
	}
}

func TestStoreErrorsReachHandlerWithoutFailingResponse(t *testing.T) {
	wantErr := errors.New("store unavailable")
	st := &mockStore{addErr: wantErr}
	var gotErr error
	wrapped := WrapWithLimitsAndErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), st, false, false, nil, 1024, 1, func(err error) {
		gotErr = err
	})
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("error handler got %v, want wrapped %v", gotErr, wantErr)
	}
}

func TestErrorHandlerPanicDoesNotReplaceApplicationResult(t *testing.T) {
	st := &mockStore{addErr: errors.New("store unavailable")}
	onError := func(error) { panic("error callback failed") }

	t.Run("normal response", func(t *testing.T) {
		wrapped := WrapWithLimitsAndErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}), st, false, false, nil, 1024, 1, onError)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
		}
	})

	t.Run("application panic", func(t *testing.T) {
		wrapped := WrapWithLimitsAndErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("application failed")
		}), st, false, false, nil, 1024, 1, onError)
		defer func() {
			if recovered := recover(); recovered != "application failed" {
				t.Fatalf("recovered %v, want original application panic", recovered)
			}
		}()
		wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestProfilingTraceDurationUnits(t *testing.T) {
	st := &mockStore{}
	profiler := profiling.NewProfiler(10)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := GetTracer(r.Context())
		if tracer == nil {
			t.Fatal("profiling middleware did not attach tracer")
		}
		tracer.RecordHTTP(http.MethodGet, "http://example.test", 2*time.Millisecond, http.StatusOK, nil)
		time.Sleep(2 * time.Millisecond)
	})

	wrapped := WrapWithProfilingAndLimits(handler, st, false, false, nil, profiler, 1024, 1)
	wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if len(st.logs) != 1 || len(st.logs[0].MiddlewareTrace) != 1 {
		t.Fatalf("missing middleware trace: %+v", st.logs)
	}
	root := st.logs[0].MiddlewareTrace[0]
	duration, ok := root["duration"].(int64)
	if !ok {
		t.Fatalf("root duration has type %T, want int64", root["duration"])
	}
	if duration < 1 || duration > 1000 {
		t.Fatalf("root duration = %d, want milliseconds", duration)
	}
	children, ok := root["children"].([]TraceEntry)
	if !ok || len(children) != 1 {
		t.Fatalf("unexpected child traces: %#v", root["children"])
	}
	if children[0].Duration != 2*time.Millisecond {
		t.Fatalf("child duration = %d, want %d ns", children[0].Duration, 2*time.Millisecond)
	}
}
