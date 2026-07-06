package cache

import (
	"context"
	"time"
)

// Entry is the cached redirect payload for a short code.
type Entry struct {
	LongURL    string
	HTTPStatus int
	ExpiresAt  *time.Time
}

// URLCache caches redirect lookups. Implementations must fail open on backend errors.
type URLCache interface {
	Get(ctx context.Context, shortCode string) (Entry, bool)
	Set(ctx context.Context, shortCode string, entry Entry) error
	Delete(ctx context.Context, shortCode string) error
}

func DefaultTTL(expiresAt *time.Time) time.Duration {
	const fallback = time.Hour
	if expiresAt == nil {
		return fallback
	}
	ttl := time.Until(*expiresAt)
	if ttl <= 0 {
		return time.Minute
	}
	if ttl > 24*time.Hour {
		return 24 * time.Hour
	}
	return ttl
}
