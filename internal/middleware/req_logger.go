package middleware

import (
	"net/http"
	"time"

	"com.perkunas/internal/logger"
)

type wrappedResponseWritter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedResponseWritter) WriteHEader(statusCcode int) {
	w.ResponseWriter.WriteHeader(statusCcode)
	w.statusCode = statusCcode
}

func LogReq(next http.Handler) http.Handler {
	log := logger.WithJSONFormat()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// status code is not directly available on w http.ResponseWriter, so we need a wrapper to capture it
		wr := &wrappedResponseWritter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wr, r)

		end := time.Since(start).Seconds()
		log.Info("req", "method", r.Method, "status", wr.statusCode, "path", r.URL.Path, "caller", r.RemoteAddr, "duration", end)
	})
}
