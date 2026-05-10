package middleware

import (
	"bytes"
	"net/http"

	"Finance-Manager-System/internal/infrastructure/cache"

	"go.uber.org/zap"
)

type cacheResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *cacheResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func CacheHTTPMiddleware(redisCache *cache.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redisCache == nil || !redisCache.Enabled() {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := GetUserID(r.Context())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodGet {
				key := redisCache.BuildRequestKey(userID, r.Method, r.URL.Path, r.URL.RawQuery)
				payload, found, getErr := redisCache.GetResponse(r.Context(), key)
				if getErr != nil {
					zap.L().Warn("redis_cache_get_failed", zap.String("key", key), zap.Error(getErr))
				}
				if found {
					if payload.ContentType != "" {
						w.Header().Set("Content-Type", payload.ContentType)
					}
					w.Header().Set("X-Cache", "HIT")
					w.WriteHeader(payload.StatusCode)
					_, _ = w.Write([]byte(payload.Body))
					return
				}

				crw := &cacheResponseWriter{ResponseWriter: w}
				next.ServeHTTP(crw, r)
				if crw.status == 0 {
					crw.status = http.StatusOK
				}
				if crw.status == http.StatusOK {
					setErr := redisCache.SetResponse(r.Context(), key, cache.ResponsePayload{
						StatusCode:  crw.status,
						ContentType: w.Header().Get("Content-Type"),
						Body:        crw.body.String(),
					})
					if setErr != nil {
						zap.L().Warn("redis_cache_set_failed", zap.String("key", key), zap.Error(setErr))
					}
				}
				return
			}

			crw := &cacheResponseWriter{ResponseWriter: w}
			next.ServeHTTP(crw, r)
			if crw.status == 0 {
				crw.status = http.StatusOK
			}
			if crw.status >= 200 && crw.status < 300 {
				if err := redisCache.InvalidateByUser(r.Context(), userID); err != nil {
					zap.L().Warn("redis_cache_invalidate_failed", zap.String("user_id", userID.String()), zap.Error(err))
				}
			}
		})
	}
}
