# ADR 0002: Sharding key — hash vs range

- **Status:** Proposed (not yet implemented)
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** README §Storage Layer & Sharding, §Benchmarking Plan (Scenario D), ADR [0001](./0001-snowflake-vs-uuid-vs-db-sequence.md)

## Context

Beyond a single Postgres primary we plan to shard `url_mappings`. Two routing
strategies are on the table: consistent hashing on `short_code`, or range
partitioning on the Snowflake ID. The choice affects write distribution,
redirect latency, and the cost of adding or removing shards.

The project's explicit goal is to *measure* the trade-off, not just pick one —
the README benchmark table requires a sharded vs non-sharded comparison.

## Decision

*Proposed, pending the Phase 4/5 benchmark.* The default plan:

- Use **hash on `short_code`** (consistent-hash ring) for the redirect-optimised
  deployment: even read spread, O(1) routing.
- Keep **range on Snowflake ID** available as the alternative for write-heavy
  benchmark runs.
- Put both behind a single router abstraction so both modes are reproducible on
  the same hardware.

## Consequences

### Positive

- Produces the empirical sharded-vs-non-sharded numbers the project exists to show.
- Hash routing spreads the hot read path evenly and avoids hot shards on viral links.

### Negative / costs

- Hash routing needs consistent hashing (or a proxy) to avoid remapping most keys
  when the shard count changes.
- Range routing can create hot shards if access skews toward recently created links.
- Sharding adds operational complexity (per-shard migrations, connection pools)
  and makes cross-shard queries painful.

## Alternatives considered

| Option | Why not |
|--------|---------|
| No sharding (single primary + read replicas) | May well be sufficient; the benchmark exists to prove or disprove it before adding complexity. |
| Vitess / ProxySQL-style proxy sharding | Heavier dependency to learn; app-level routing is the smaller first step. |
| Shard by `user_id` | Skews badly for anonymous links and concentrates viral public links. |

## References

- Consistent Hashing — https://en.wikipedia.org/wiki/Consistent_hashing
- README §Benchmarking Plan
