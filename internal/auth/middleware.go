package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JathinShyam/NexusLink/internal/httpx"
)

// SessionValidator checks that a signed token still maps to an active
// server-side session. It lets logout revoke tokens before they expire.
type SessionValidator interface {
	ValidateSession(ctx context.Context, userID, sessionID string) error
}

// Middleware requires a valid bearer token. When a SessionValidator is passed,
// tokens that carry a session ID are also checked against the session store.
func Middleware(jwtService *JWT, sessions ...SessionValidator) func(http.Handler) http.Handler {
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

			if err := validateSession(r.Context(), sessions, claims); err != nil {
				writeSessionError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithSession(r.Context(), claims.UserID, claims.SessionID)))
		})
	}
}

// OptionalMiddleware attaches the user when a valid token is present. A token
// that is present but invalid is rejected instead of silently creating an
// anonymous link, so stale sessions surface to the client.
func OptionalMiddleware(jwtService *JWT, sessions ...SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
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

			if err := validateSession(r.Context(), sessions, claims); err != nil {
				writeSessionError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithSession(r.Context(), claims.UserID, claims.SessionID)))
		})
	}
}

func validateSession(ctx context.Context, sessions []SessionValidator, claims *Claims) error {
	if claims.SessionID == "" || len(sessions) == 0 || sessions[0] == nil {
		return nil
	}
	return sessions[0].ValidateSession(ctx, claims.UserID, claims.SessionID)
}

func writeSessionError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrSessionInvalid) {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "session expired; sign in again")
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to validate session")
}
