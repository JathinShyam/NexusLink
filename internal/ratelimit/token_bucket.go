package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  ts = now
end

local delta = math.max(0, now - ts)
tokens = math.min(capacity, tokens + (delta * refill_rate))
ts = now

if tokens < cost then
  redis.call("HMSET", key, "tokens", tokens, "ts", ts)
  redis.call("PEXPIRE", key, 120000)
  return 0
end

tokens = tokens - cost
redis.call("HMSET", key, "tokens", tokens, "ts", ts)
redis.call("PEXPIRE", key, 120000)
return 1
`)

// TokenBucketLimiter allows steady traffic with controlled bursts.
type TokenBucketLimiter struct {
	client     *redis.Client
	prefix     string
	capacity   int64
	refillRate float64
}

func NewTokenBucketLimiter(client *redis.Client, prefix string, capacity int64, refillPerSecond float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		client:     client,
		prefix:     prefix,
		capacity:   capacity,
		refillRate: refillPerSecond,
	}
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
	now := time.Now().UnixMilli()

	result, err := tokenBucketScript.Run(
		ctx,
		l.client,
		[]string{redisKey},
		now,
		l.refillRate,
		l.capacity,
		1,
	).Int()
	if err != nil {
		return true, nil
	}
	return result == 1, nil
}
