package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	sessionIDKey contextKey = "sessionID"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrMissingToken = errors.New("missing token")
	// ErrSessionInvalid marks a token that no longer maps to an active
	// server-side session, as opposed to a datastore failure.
	ErrSessionInvalid = errors.New("invalid session")
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret string, ttl time.Duration) *JWT {
	return &JWT{secret: []byte(secret), ttl: ttl}
}

func (j *JWT) TTL() time.Duration {
	return j.ttl
}

// Issue creates a token without a server-side session. It is kept for the
// development token helper and backwards compatibility.
func (j *JWT) Issue(userID string) (string, error) {
	return j.IssueSession(userID, "")
}

// IssueSession creates a token tied to a revocable session ID.
func (j *JWT) IssueSession(userID, sessionID string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ID:        sessionID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWT) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func WithSession(ctx context.Context, userID, sessionID string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	if sessionID != "" {
		ctx = context.WithValue(ctx, sessionIDKey, sessionID)
	}
	return ctx
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok && sessionID != ""
}
