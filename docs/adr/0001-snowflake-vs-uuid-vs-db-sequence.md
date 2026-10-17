# ADR 0001: Distributed IDs — Snowflake vs UUID vs DB sequence

- **Status:** Accepted
- **Date:** 2026-10-18
- **Deciders:** NexusLink maintainers
- **Related:** [`internal/idgen/snowflake.go`](../../internal/idgen/snowflake.go), [`internal/encoding/base62`](../../internal/encoding/base62), README §Core Components, ADR [0002](./0002-hash-sharding-vs-range-sharding.md)

## Context

Every shorten request needs a unique, non-guessable value that we encode into a
short code. The generator must work across multiple stateless API instances,
avoid a single point of contention on the write path, and stay compact once
base62-encoded (target 6–8 characters).

Constraints:

- Horizontal scale of the shorten path (NF1) — no central lock.
- Non-guessable codes to resist enumeration.
- Ideally time-sortable, so the future sharding work (ADR 0002) has range options.

## Decision

Generate 64-bit **Snowflake-style** IDs and base62-encode them:

```
41 bits timestamp (ms since epoch 2021-01-01 UTC)
10 bits machine ID (0–1023)
12 bits per-millisecond sequence
```

- Each instance sets a unique `SNOWFLAKE_MACHINE_ID`.
- `Next()` is mutex-guarded; if the clock moves backwards it waits until the
  timestamp advances rather than emitting duplicates.
- The DB-sequence generator (`internal/idgen/sequence.go`) is retained for tests
  and local simplicity, but production wiring uses Snowflake.

## Consequences

### Positive

- No coordination between instances; no hot central sequence to contend on.
- Roughly time-ordered IDs → good locality for range-style access patterns.
- IDs are globally unique, so base62 collisions are not a design concern.
- 12 sequence bits allow 4096 IDs/ms per machine — far above expected shorten QPS.

### Negative / costs

- Requires a unique machine ID per instance; a misconfiguration is a correctness
  risk (mitigated by validation in `NewSnowflakeGenerator`).
- Depends on reasonably accurate clocks (NTP); small skew is handled by waiting.
- Fixed ceiling of 1024 machines; a redesign is needed beyond that scale.

## Alternatives considered

| Option | Why not |
|--------|---------|
| Postgres serial / sequence | Central contention on the hot write path; harder to shard; leaked ordering reveals volume. |
| UUIDv4 | Random, non-sortable, and encodes to a long, ugly short link. |
| UUIDv7 (time-ordered) | Viable, but 128 bits encodes longer than needed; Snowflake gives the same benefits at 64 bits. |

## References

- Twitter Snowflake — https://github.com/twitter-archive/snowflake
- Base62 alphabet: `0-9a-zA-Z` (`internal/encoding/base62`)
