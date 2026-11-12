package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeSessionValidator struct {
	err       error
	called    bool
	userID    string
	sessionID string
}

func (f *fakeSessionValidator) ValidateSession(_ context.Context, userID, sessionID string) error {
	f.called = true
	f.userID = userID
	f.sessionID = sessionID
	return f.err
}

func bearerRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestMiddlewareSessionValidation(t *testing.T) {
	jwtService := NewJWT("test-secret", time.Hour)
	token, err := jwtService.IssueSession("usr_1", "ses_1")
	if err != nil {
		t.Fatalf("IssueSession() error = %v", err)
	}

	cases := []struct {
		name       string
		validator  *fakeSessionValidator
		wantStatus int
	}{
		{
			name:       "active session",
			validator:  &fakeSessionValidator{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "revoked session",
			validator:  &fakeSessionValidator{err: ErrSessionInvalid},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "datastore failure",
			validator:  &fakeSessionValidator{err: errors.New("database unavailable")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := Middleware(jwtService, tc.validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := UserIDFromContext(r.Context())
				if !ok || userID != "usr_1" {
					t.Errorf("UserIDFromContext() = %q, %v", userID, ok)
				}
				sessionID, ok := SessionIDFromContext(r.Context())
				if !ok || sessionID != "ses_1" {
					t.Errorf("SessionIDFromContext() = %q, %v", sessionID, ok)
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, bearerRequest(token))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.validator.userID != "usr_1" || tc.validator.sessionID != "ses_1" {
				t.Errorf("validator got user=%q session=%q", tc.validator.userID, tc.validator.sessionID)
			}
		})
	}
}

func TestMiddlewareRejectsMissingToken(t *testing.T) {
	handler := Middleware(NewJWT("test-secret", time.Hour))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not run")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOptionalMiddlewareWithoutToken(t *testing.T) {
	handler := OptionalMiddleware(NewJWT("test-secret", time.Hour))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserIDFromContext(r.Context()); ok {
			t.Error("anonymous request unexpectedly had a user")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/shorten", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestOptionalMiddlewareRejectsInvalidToken(t *testing.T) {
	handler := OptionalMiddleware(NewJWT("test-secret", time.Hour))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not run")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, bearerRequest("not-a-jwt"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSessionlessTokenSkipsValidation(t *testing.T) {
	jwtService := NewJWT("test-secret", time.Hour)
	token, err := jwtService.Issue("usr_legacy")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	validator := &fakeSessionValidator{err: ErrSessionInvalid}
	handler := Middleware(jwtService, validator)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, bearerRequest(token))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if validator.called {
		t.Error("validator should not run for a sessionless token")
	}
}
