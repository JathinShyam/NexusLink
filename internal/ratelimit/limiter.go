package ratelimit

import "context"

// Limiter decides whether a request identified by key is allowed.
// Implementations must fail open when the backend is unavailable.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
