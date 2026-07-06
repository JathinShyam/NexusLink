package auth

import (
	"net/http"
	"strings"

	"github.com/JathinShyam/NexusLink/internal/httpx"
)

func Middleware(jwtService *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
				return
			}

			claims, err := jwtService.Parse(parts[1])
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), claims.UserID)))
		})
	}
}

func OptionalMiddleware(jwtService *JWT) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if claims, err := jwtService.Parse(parts[1]); err == nil {
					r = r.WithContext(WithUserID(r.Context(), claims.UserID))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
