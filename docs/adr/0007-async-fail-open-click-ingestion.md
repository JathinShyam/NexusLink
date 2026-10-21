# ADR 0007: Async, fail-open click ingestion with at-least-once delivery

- **Status:** Accepted
- **Date:** 2026-10-21
- **Deciders:** NexusLink maintainers
- **Related:** [`internal/analytics/async.go`](../../internal/analytics/async.go), [`internal/analytics/kafka.go`](../../internal/analytics/kafka.go), [`docs/runbook.md`](../runbook.md), ADR 0005

## Context

Redirects are the critical path. Analytics is valuable but must never become a
reason a short link fails to resolve (NF3). The broker can be slow, unreachable,
or misconfigured, and the redirect path handles thousands of requests per
second.

Two failure postures are possible:

- **Synchronous produce** — the handler waits for Kafka before returning the
  redirect. Simple and lossless, but broker latency (and outages) become
  redirect latency (and errors).
- **Decoupled produce** — the handler hands the event to an in-process buffer
  that a background worker drains. Fast and fail-open, but events can be lost if
  the buffer overflows or the process dies before draining.

Exactly-once delivery would require transactional produce/consume or a dedupe
store on the hot path; the product does not need it.

## Decision

**Decouple and fail open.** The redirect handler only calls
`analytics.Publisher.Publish`, which enqueues onto a bounded channel and never
blocks:

- Buffer size is configurable (`ANALYTICS_BUFFER_SIZE`, default `1024`).
- When the buffer is full, the event is **dropped**, a warning is logged, and
  `click_events_dropped_total` is incremented.
- A single background goroutine drains the buffer and calls the configured
  `Sink` with a 5s per-event timeout. Sink errors increment
  `click_events_errors_total`; redirects are unaffected.
- Analytics is **disabled by default** (`ANALYTICS_ENABLED=false`) so the
  behaviour and dependencies of the redirect path do not change unless opted in.

Delivery to Kafka is **at-least-once**:

- Records are JSON, keyed by `short_code`, so clicks for one link are ordered
  within a partition.
- The producer is idempotent (`franz-go` default), uses `acks=all`, and retries
  transient failures.
- Each event carries a `request_id` so the future consumer can deduplicate
  replays (`INSERT` with `ReplacingMergeTree`/dedupe key) and tolerate
  at-least-once delivery.

At-least-once plus an idempotent consumer was chosen over exactly-once because
it keeps the hot path cheap and the operational model simple; approximate counts
are acceptable for analytics.

## Consequences

### Positive

- Broker latency or an outage cannot slow down or fail redirects.
- Backpressure is explicit and observable (`click_events_dropped_total`), not a
  hidden queue that grows until OOM.
- The `Sink` seam allows a log sink locally and Kafka (or another backend) in
  production without touching the handler.
- Keying by short code preserves per-link ordering and lets the consumer scale
  by partition later.

### Negative

- Events can be lost on buffer overflow, shutdown, or crash; analytics is
  best-effort, not billing-grade.
- At-least-once delivery means duplicates are possible and the consumer must
  deduplicate by `request_id` (or accept over-counts).
- `acks=all`, retries, and per-event timeouts add producer latency — hidden
  behind the async buffer, but they bound throughput when the broker is slow.

### Neutral

- Redis-style fail-open (ADR 0005) now extends to the analytics sink: degraded
  analytics is preferred over failed redirects.
- The consumer and ClickHouse schema are still to be built; this ADR only fixes
  the ingestion contract.
