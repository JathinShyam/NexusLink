# ADR 0006: Default redirect status — 302 vs 301

- **Status:** Accepted
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** [`internal/redirect/resolver.go`](../../internal/redirect/resolver.go), [`internal/shorten/service.go`](../../internal/shorten/service.go), `migrations/001_url_mappings.up.sql`

## Context

`GET /{short_code}` must redirect to the stored long URL. HTTP offers several
redirect statuses; the relevant pair is:

- **301 Moved Permanently** — aggressively cached by browsers and intermediaries.
- **302 Found** — not permanently cached; the client re-requests the short link.

The project needs every click to reach the service so it can later emit a click
event to the analytics pipeline (Phase 4).

## Decision

Default to **302 Found**, stored per-row in `url_mappings.http_status` (default
`302`). The redirect handler uses the stored status, falling back to `302` when
the cached or persisted value is unset.

301 remains possible for links that are explicitly permanent, since the column is
per-link.

## Consequences

### Positive

- Each click hits the service, so click tracking and expiry checks always run.
- Changing a link's destination takes effect immediately (no stuck client cache).
- Soft-deleted links correctly return 404 instead of a stale cached redirect.

### Negative / costs

- Slightly less efficient: no client/intermediary caching of the redirect.
- More load than a 301-based design would generate (acceptable at our scale and
  required for analytics).

## Alternatives considered

| Option | Why not |
|--------|---------|
| 301 by default | Browsers cache it, so clicks and expiry are missed and destination changes don't propagate. |
| 307 / 308 | Preserve method/body semantics we don't need for a simple GET redirect; 302 is the least surprising. |

## References

- MDN — 302 Found vs 301 Moved Permanently
