package middleware

import (
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func ZapRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		zap.L().Info(
			"http_request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.Int("status", ww.Status()),
			zap.Int("bytes", ww.BytesWritten()),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("request_id", chiMiddleware.GetReqID(r.Context())),
		)
	})
}

func ZapRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				zap.L().Error(
					"http_panic_recovered",
					zap.Any("panic", rec),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("request_id", chiMiddleware.GetReqID(r.Context())),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
