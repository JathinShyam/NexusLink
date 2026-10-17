# ADR 0004: Rate limiting algorithms and per-endpoint choice

- **Status:** Accepted
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** [`internal/ratelimit`](../../internal/ratelimit), [`internal/config`](../../internal/config/config.go), ADR [0005](./0005-fail-open-vs-fail-closed-on-redis-outage.md)

## Context

The API needs abuse protection scoped per IP / per user / per endpoint (NF4),
enforced atomically across stateless instances. Several algorithms trade memory,
precision, and burst behaviour. The project's learning goal is to implement more
than one and justify a default per endpoint.

## Decision

Implement two algorithms and ship one per endpoint:

| Endpoint | Algorithm | Defaults | Why |
|----------|-----------|----------|-----|
| `POST /api/v1/shorten` | **Fixed window** (`INCR` + `EXPIRE`) | 30 req / 60 s | Writes are bursty but low-volume; a simple per-IP cap is enough. |
| `GET /{short_code}` | **Token bucket** (Redis Lua) | capacity 100, refill 50/s | Read-heavy; absorbs viral bursts while holding a steady rate. |

Shared details:

- Key is `ClientIP + ":" + endpoint`; client IP is the first `X-Forwarded-For`
  hop, falling back to `RemoteAddr`.
- Check-and-increment runs inside Redis (Lua) for atomicity; behaviour is
  **fail-open** (see ADR 0005).
- All thresholds are configurable (`RATE_LIMIT_*` env vars).

## Consequences

### Positive

- Fixed window is trivial to reason about and cheap on memory for writes.
- Token bucket smooths read traffic and tolerates legitimate bursts better.
- Exceeding a limit returns `429 RATE_LIMITED` consistently.

### Negative / costs

- Fixed window allows up to ~2× the limit at window boundaries (accepted; writes
  are low-volume).
- Two code paths to test and maintain.
- Fail-open means the limit is not enforced at all while Redis is down.

## Alternatives considered

| Option | Why not |
|--------|---------|
| Sliding window counter | Better boundary behaviour, more state; not needed for the low-volume write path. |
| Sliding window log | Precise but memory-heavy per client; overkill here. |
| Token bucket everywhere | Reasonable, but adds moving parts to a path where a fixed cap is clearer. |
| In-process limiter only | Not shared across instances; wrong for a horizontally scaled service. |

## References

- Redis rate limiting patterns — https://redis.io/docs/latest/develop/use/patterns/rate-limiting/
