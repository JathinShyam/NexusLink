package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyPrefix = "url:"

type redisEntry struct {
	LongURL    string     `json:"long_url"`
	HTTPStatus int        `json:"http_status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, shortCode string) (Entry, bool) {
	data, err := c.client.Get(ctx, keyPrefix+shortCode).Bytes()
	if err != nil {
		return Entry{}, false
	}

	var cached redisEntry
	if err := json.Unmarshal(data, &cached); err != nil {
		return Entry{}, false
	}

	if cached.ExpiresAt != nil && cached.ExpiresAt.Before(time.Now().UTC()) {
		_ = c.client.Del(ctx, keyPrefix+shortCode).Err()
		return Entry{}, false
	}

	return Entry(cached), true
}

func (c *RedisCache) Set(ctx context.Context, shortCode string, entry Entry) error {
	payload, err := json.Marshal(redisEntry(entry))
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	ttl := DefaultTTL(entry.ExpiresAt)
	if err := c.client.Set(ctx, keyPrefix+shortCode, payload, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) Delete(ctx context.Context, shortCode string) error {
	if err := c.client.Del(ctx, keyPrefix+shortCode).Err(); err != nil {
		return err
	}
	return nil
}
