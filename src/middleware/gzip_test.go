package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip_CompressesResponseWhenClientAcceptsGzip(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding to be gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
		t.Errorf("expected Vary header to include Accept-Encoding")
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}

	if string(body) != "Hello, World!" {
		t.Errorf("expected body to be 'Hello, World!', got '%s'", string(body))
	}
}

func TestGzip_DoesNotCompressWhenClientDoesNotAcceptGzip(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected response to not be gzip encoded")
	}

	if rec.Body.String() != "Hello, World!" {
		t.Errorf("expected body to be 'Hello, World!', got '%s'", rec.Body.String())
	}
}
