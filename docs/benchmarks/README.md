# Benchmarks

Load-test methodology and results. **Nothing measured yet** — this folder is
scaffolding for Phase 5 (observability & hardening).

## Scenarios (from the README)

| Scenario | Description | Primary metrics |
|----------|-------------|-----------------|
| A — Baseline | Single app instance, no shard | shorten + redirect p50/p99, RPS |
| B — Read heavy | 95% redirect, 5% shorten; one viral `short_code` | cache hit %, p99 |
| C — Write burst | Many unique shortens | DB CPU, insert TPS |
| D — Sharded vs not | Same hardware, scale DB/app | throughput, p99, ops complexity |

## Tooling

- **k6** or **Locust** for HTTP load.
- **Prometheus** scraping `/metrics` during runs (QPS, p99, cache hit %, errors).
- Export Grafana screenshots here when dashboards exist.

## Results template (fill after Phase 4/5)

| Config | Redirect RPS | Shorten RPS | p99 redirect (ms) | Cache hit % | Notes |
|--------|--------------|-------------|-------------------|-------------|-------|
| Single PG, no Redis | | | | N/A | |
| PG + Redis | | | | | |
| Sharded PG (N shards) | | | | | |
| Non-sharded PG (same total CPU) | | | | | |

> Replace this table with real data before publishing. The README narrative
> sentence about sharded vs non-sharded throughput should cite these numbers.
