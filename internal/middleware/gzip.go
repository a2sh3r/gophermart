package middleware

import (
	"compress/gzip"
	"github.com/a2sh3r/gophermart/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func WithGzip() func(next http.Handler) http.Handler {
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

				defer func() {
					if err := gz.Close(); err != nil {
						logger.Log.Error("Failed to close gzip body", zap.Error(err))
					}
				}()

				w = gzipResponseWriter{Writer: gz, ResponseWriter: w}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}
