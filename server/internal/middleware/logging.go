package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			writer := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(writer, r)

			status := writer.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.InfoContext(
				r.Context(),
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"bytes", writer.bytes,
				"duration_ms", time.Since(started).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
