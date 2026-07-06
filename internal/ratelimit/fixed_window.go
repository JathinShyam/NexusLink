package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// FixedWindowLimiter counts requests per time window (simple, boundary spikes possible).
type FixedWindowLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
	prefix string
}

func NewFixedWindowLimiter(client *redis.Client, prefix string, limit int64, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		client: client,
		limit:  limit,
		window: window,
		prefix: prefix,
	}
}

func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return true, nil
	}
	if count == 1 {
		_ = l.client.Expire(ctx, redisKey, l.window).Err()
	}
	return count <= l.limit, nil
}
