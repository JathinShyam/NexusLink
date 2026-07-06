# Phase 3: Scale Core (Cache, Rate Limit, Auth)

## Goal

Make the MVP more production-like: faster reads, safer traffic handling, basic user boundaries.

## Duration

2-3 weeks

## Learning Focus

- Redis caching patterns (read-through, invalidation)
- Rate limiting algorithms
- JWT auth basics
- Metrics instrumentation

## Build Tasks

- [x] Integrate Redis for redirect cache.
- [x] Add cache read-through on redirect.
- [x] Add cache update/invalidation on shorten or delete.
- [x] Implement rate limiting (start with fixed window, then token bucket).
- [x] Add JWT auth for link management endpoints.
- [x] Add Prometheus metrics for latency, hit/miss, rate-limited requests.

## Deliverables

- Visible cache hit/miss behavior.
- 429 responses when limits are exceeded.
- Auth-protected route working.

## Exit Criteria

- Redirect path is faster on cache hit than DB hit.
- You can explain why one rate limiter was chosen for default.

## Suggested Weekly Split

- Week 1: Redis cache integration
- Week 2: rate limiting algorithms
- Week 3: auth + metrics + polish
