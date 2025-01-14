package middleware

import (
	"net/http"
	"time"

	"com.perkunas/internal/logger"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func LogReq(next http.Handler) http.Handler {
	log := logger.WithJSONFormat()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wr := &wrappedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wr, r)

		end := time.Since(start).Seconds()
		log.Info("req", "method", r.Method, "status", wr.statusCode, "path", r.URL.Path, "caller", r.RemoteAddr, "duration", end)
	})
}
