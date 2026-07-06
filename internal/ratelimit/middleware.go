package ratelimit

import (
	"net"
	"net/http"
	"strings"

	"github.com/JathinShyam/NexusLink/internal/httpx"
	"github.com/JathinShyam/NexusLink/internal/metrics"
)

func ClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func Middleware(limiter Limiter, endpoint string, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ClientIP(r) + ":" + endpoint
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
