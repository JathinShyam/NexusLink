# ADR 0003: Analytics store — ClickHouse vs Postgres

- **Status:** Proposed (Phase 4, not yet implemented)
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** README §Analytics Pipeline, `roadmap/phases/04-analytics-pipeline.md`

## Context

Redirects must not block on analytics (NF3). At design scale that is ~100M
click events/day (~20 GB/day raw), which we intend to store for per-link
aggregations: totals, clicks over time, top referrers, and geo distribution.

The question is where the click events land once consumed from Kafka.

## Decision

*Proposed:* consume `url_clicks` from Kafka and batch-insert into **ClickHouse**
(`clicks` table, `PARTITION BY toYYYYMM(timestamp)`, `ORDER BY (short_code, timestamp)`),
serving the analytics API from it.

As an intermediate learning step, **Postgres + materialized views** may be used
first so the pipeline (producer → consumer → API → dashboard) can be validated
before introducing ClickHouse.

## Consequences

### Positive

- Columnar storage and compression are far cheaper for billions of append-only rows.
- Fast `GROUP BY`-heavy dashboard queries via materialized views.
- TTL + monthly partitioning make retention/rollups straightforward.

### Negative / costs

- A second datastore to run, back up, and monitor.
- Not a great fit for point updates; we treat click data as append-only.
- At-least-once Kafka delivery means the consumer must be idempotent (dedupe by
  `request_id` or accept approximate counts for learning).

## Alternatives considered

| Option | Why not |
|--------|---------|
| Postgres only | Works to modest scale, but OLTP is a poor fit for billion-row aggregations and bloats the primary. |
| ClickHouse only from day one | Skips a simpler validation step; adds infra before the pipeline is proven. |
| Pre-aggregate in Redis counters | Good for a cheap "total clicks" number, but loses referrer/geo/time-series detail. |

## References

- ClickHouse docs — https://clickhouse.com/docs
- README §Data Model (ClickHouse `clicks` schema)
