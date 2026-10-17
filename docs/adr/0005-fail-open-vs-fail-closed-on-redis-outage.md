# ADR 0005: Fail-open vs fail-closed when Redis is unavailable

- **Status:** Accepted
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** [`internal/cache/noop.go`](../../internal/cache/noop.go), [`internal/ratelimit/limiter.go`](../../internal/ratelimit/limiter.go), README §Failure Scenarios

## Context

Redis backs both the redirect cache and the rate limiters. If Redis becomes
unavailable, the service must decide between two failure postures:

- **Fail-closed** — reject requests (or block redirects) until Redis returns.
- **Fail-open** — keep serving, degrade gracefully, and alert.

For a URL shortener, redirects are the critical path: a dead redirect is a
broken link for every user.

## Decision

**Fail-open everywhere** for Redis-backed components:

- **Cache:** on connect failure at startup we fall back to `NoopCache` (always a
  miss), so every redirect reads Postgres. `Get`/`Set` errors are treated as a
  miss / no-op.
- **Rate limiting:** if the Redis command (or Lua script) errors, `Allow`
  returns `(true, nil)` — the request is allowed and the limit is simply not
  enforced.

The trade-off is deliberate: **availability over strict enforcement**. The
degradation is observable via `cache_misses_total` climbing and (optionally)
Redis error logs, and should be paired with an alert.

## Consequences

### Positive

- A Redis outage does not take down redirects or shorten.
- No thundering-herd of 5xx errors; the system degrades to database-only reads.

### Negative / costs

- Latency rises while Redis is down (every redirect hits Postgres, plus rate
  limiting is effectively disabled).
- The service is briefly exposed to abuse during an outage.
- Fail-open must never be assumed safe silently — it requires monitoring.

## Alternatives considered

| Option | Why not |
|--------|---------|
| Fail-closed on rate limit | Turns a Redis blip into a full outage; unacceptable for the redirect path. |
| Fail-closed on cache, fail-open on limits | Cache misses are already tolerated via Postgres; no reason to reject. |
| Local in-process limiter as fallback | Approximate and per-instance, but a reasonable future enhancement behind a feature flag. |

## References

- README §Failure Scenarios (Redis unavailable)
- `internal/cache/cache.go` — "Implementations must fail open on backend errors."
