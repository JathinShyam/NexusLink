package redirect

import (
	"context"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/cache"
	"github.com/JathinShyam/NexusLink/internal/metrics"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type Resolver struct {
	repo    storage.Repository
	cache   cache.URLCache
	metrics *metrics.Metrics
}

func NewResolver(repo storage.Repository, urlCache cache.URLCache, m *metrics.Metrics) *Resolver {
	return &Resolver{repo: repo, cache: urlCache, metrics: m}
}

type resolved struct {
	longURL    string
	httpStatus int
}

func (r *Resolver) Resolve(ctx context.Context, shortCode string) (resolved, error) {
	if entry, ok := r.cache.Get(ctx, shortCode); ok {
		if r.metrics != nil {
			r.metrics.CacheHit()
		}
		if entry.ExpiresAt != nil && entry.ExpiresAt.Before(time.Now().UTC()) {
			_ = r.cache.Delete(ctx, shortCode)
			return resolved{}, storage.ErrNotFound
		}
		status := entry.HTTPStatus
		if status == 0 {
			status = http.StatusFound
		}
		return resolved{longURL: entry.LongURL, httpStatus: status}, nil
	}

	if r.metrics != nil {
		r.metrics.CacheMiss()
	}

	mapping, err := r.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return resolved{}, err
	}

	if !mapping.IsActive {
		return resolved{}, storage.ErrNotFound
	}
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now().UTC()) {
		return resolved{}, storage.ErrNotFound
	}

	status := mapping.HTTPStatus
	if status == 0 {
		status = http.StatusFound
	}

	_ = r.cache.Set(ctx, shortCode, cache.Entry{
		LongURL:    mapping.LongURL,
		HTTPStatus: status,
		ExpiresAt:  mapping.ExpiresAt,
	})

	return resolved{longURL: mapping.LongURL, httpStatus: status}, nil
}

func (r *Resolver) Invalidate(ctx context.Context, shortCode string) {
	_ = r.cache.Delete(ctx, shortCode)
}
