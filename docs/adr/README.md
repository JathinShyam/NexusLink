# Architecture Decision Records

An ADR captures a single significant decision: the context, the choice, and the
trade-offs we accept. Records are immutable once **Accepted** — to change a
decision, add a new ADR that **Supersedes** the old one rather than editing it.

## Index

| # | Decision | Status |
|---|----------|--------|
| [0001](./0001-snowflake-vs-uuid-vs-db-sequence.md) | Distributed IDs: Snowflake vs UUID vs DB sequence | Accepted |
| [0002](./0002-hash-sharding-vs-range-sharding.md) | Sharding key: hash vs range | Proposed |
| [0003](./0003-clickhouse-vs-postgres-for-analytics.md) | Analytics store: ClickHouse vs Postgres | Proposed |
| [0004](./0004-rate-limiting-algorithms.md) | Rate limiting algorithms and per-endpoint choice | Accepted |
| [0005](./0005-fail-open-vs-fail-closed-on-redis-outage.md) | Fail-open vs fail-closed when Redis is unavailable | Accepted |
| [0006](./0006-302-vs-301-default-redirect.md) | Default redirect status: 302 vs 301 | Accepted |

## Status legend

- **Proposed** — written down, not yet implemented.
- **Accepted** — implemented and in use.
- **Superseded by ADR-NNNN** — replaced; kept for history.
- **Deprecated** — no longer relevant.

## Adding a record

1. Copy [`0000-adr-template.md`](./0000-adr-template.md) to `NNNN-short-title.md`
   using the next free number.
2. Fill in context, decision, and consequences.
3. Add a row to the index above.
