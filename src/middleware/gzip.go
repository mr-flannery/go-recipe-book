package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// WebSocket upgrades require hijacking the raw TCP connection.
		// Wrapping the ResponseWriter in gzipResponseWriter would hide the
		// http.Hijacker interface and cause the upgrade to fail.
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)
		gz.Reset(w)

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.Header().Add("Vary", "Accept-Encoding")

		grw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
		next.ServeHTTP(grw, r)
		gz.Close()
	})
}
