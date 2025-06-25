package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/a2sh3r/gophermart/internal/logger"
	"go.uber.org/zap"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func NewGzipMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Encoding") == "gzip" {

				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "invalid gzip", http.StatusBadRequest)
					return
				}

				defer func() {
					if err := gz.Close(); err != nil {
						logger.Log.Error("Failed to close gzip body", zap.Error(err))
					}
				}()

				r.Body = io.NopCloser(gz)
			}

			if acceptsGzip(r) {
				w.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(w)

				grw := &gzipResponseWriter{Writer: gz, ResponseWriter: w, statusCode: 200}
				w = grw

				defer func() {
					if grw.statusCode == http.StatusNoContent || grw.statusCode == http.StatusNotModified || r.Method == http.MethodHead {
						return
					}
					if err := gz.Close(); err != nil {
						logger.Log.Error("Failed to close gzip body", zap.Error(err))
					}
				}()
			}

			next.ServeHTTP(w, r)
		})
	}
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}
