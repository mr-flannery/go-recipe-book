package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/logging"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// Hijack delegates to the underlying ResponseWriter so WebSocket upgrades work.
func (rr *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rr.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush delegates to the underlying ResponseWriter so SSE streaming works.
func (rr *responseRecorder) Flush() {
	if flusher, ok := rr.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func WideEventMiddleware(next http.Handler) http.Handler {
	envContext := logging.GetEnvContext()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		event := logging.NewWideEvent()
		event.SetMany(envContext)
		event.SetMany(map[string]any{
			"http.method":      r.Method,
			"http.path":        r.URL.Path,
			"http.query":       r.URL.RawQuery,
			"http.user_agent":  r.UserAgent(),
			"http.remote_addr": r.RemoteAddr,
		})

		if referer := r.Referer(); referer != "" {
			event.Set("http.referer", referer)
		}

		ctx := logging.ContextWithWideEvent(r.Context(), event)
		r = r.WithContext(ctx)

		recorder := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			duration := time.Since(startTime)
			event.Set("duration_ms", duration.Milliseconds())
			event.Set("http.status_code", recorder.statusCode)

			if recorder.statusCode >= 400 {
				event.Set("outcome", "error")
			} else {
				event.Set("outcome", "success")
			}

			logging.Emit(ctx)
		}()

		next.ServeHTTP(recorder, r)
	})
}
