package middleware

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// hijackableRecorder wraps httptest.ResponseRecorder and implements http.Hijacker and http.Flusher.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
	hijackCalled bool
	flushCalled  bool
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijackCalled = true
	return nil, nil, nil
}

func (h *hijackableRecorder) Flush() {
	h.flushCalled = true
	h.ResponseRecorder.Flush()
}

func TestWideEventMiddleware_HijackDelegated(t *testing.T) {
	hijackable := &hijackableRecorder{ResponseRecorder: httptest.NewRecorder()}

	var capturedWriter http.ResponseWriter
	handler := WideEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedWriter = w
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(hijackable, req)

	hj, ok := capturedWriter.(http.Hijacker)
	if !ok {
		t.Fatal("responseRecorder passed to handler does not implement http.Hijacker")
	}

	hj.Hijack()

	if !hijackable.hijackCalled {
		t.Error("Hijack() was not delegated to the underlying ResponseWriter")
	}
}

func TestWideEventMiddleware_FlushDelegated(t *testing.T) {
	hijackable := &hijackableRecorder{ResponseRecorder: httptest.NewRecorder()}

	var capturedWriter http.ResponseWriter
	handler := WideEventMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedWriter = w
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(hijackable, req)

	flusher, ok := capturedWriter.(http.Flusher)
	if !ok {
		t.Fatal("responseRecorder passed to handler does not implement http.Flusher")
	}

	flusher.Flush()

	if !hijackable.flushCalled {
		t.Error("Flush() was not delegated to the underlying ResponseWriter")
	}
}
