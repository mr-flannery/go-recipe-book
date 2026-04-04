package middleware

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// hijackableRecorder wraps httptest.ResponseRecorder and implements http.Hijacker.
type hijackableRecorder struct {
	*httptest.ResponseRecorder
	hijackCalled bool
}

func (h *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijackCalled = true
	return nil, nil, nil
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
