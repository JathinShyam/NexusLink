package ratelimit

import (
	"net/http"

	"github.com/JathinShyam/NexusLink/internal/httpx"
	"github.com/JathinShyam/NexusLink/internal/metrics"
)

func Middleware(limiter Limiter, endpoint string, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := httpx.ClientIP(r) + ":" + endpoint
			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				if m != nil {
					m.RateLimitExceeded(endpoint)
				}
				httpx.WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
