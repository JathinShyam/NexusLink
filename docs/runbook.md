# Runbook

Operational guide for running, verifying, and debugging NexusLink.

## Prerequisites

- Go 1.25+ (see `go.mod`)
- Docker & Docker Compose
- Node.js 20+ (frontend only)
- `make` (optional but assumed below)

## First run

```bash
cp .env.example .env
make docker-up      # Postgres + Redis + migrate + API + web
make health         # -> {"status":"ok"}
make ready          # -> {"status":"ready"}
```

> **Log directory permissions.** The Postgres container runs as an unprivileged
> user and writes to the bind-mounted `logs/postgres/`. If it exits immediately
> with `could not open log file ... Permission denied`, make the directory
> writable: `chmod -R 777 logs`.

## Day-to-day commands

| Command | Does |
|---------|------|
| `make docker-up` / `make docker-down` | Start / stop the full stack |
| `make run` | Run the API locally (Postgres must be reachable) |
| `make migrate-up` / `make migrate-down` | Apply / roll back migrations |
| `make health` / `make ready` | Liveness / readiness checks |
| `make demo` / `make demo-phase3` | Shorten + redirect walkthroughs |
| `make e2e` | Full automated suite against a running stack |
| `make gentoken` | Print a dev JWT |
| `make web-dev` / `make web-build` | Frontend dev server / production build |
| `make logs` / `make logs-all` / `make logs-tail` | Tail service log files |

Ports: API `:8080`, web `:3000` (Docker) or `:5173` (Vite), Postgres `:5432`,
Redis `:6379`.

## Health & observability

- `GET /health` — process is up (liveness).
- `GET /ready` — Postgres **and** Redis reachable (readiness; `503` with a
  reason when not).
- `GET /metrics` — Prometheus exposition. Key series:

| Metric | Watch for |
|--------|-----------|
| `http_requests_total{method,path,status}` | error-rate spikes (5xx) |
| `http_request_duration_seconds` | p99 redirect latency (target < 100 ms) |
| `cache_hits_total` / `cache_misses_total` | hit-ratio collapse ⇒ Redis problem |
| `rate_limit_exceeded_total{endpoint}` | a client being throttled (429s) |

## Migrations

Migrations live in `migrations/` as versioned `up`/`down` SQL pairs and are
applied by the one-shot `migrate` Compose service (or `make migrate-up`).

> After changing the `user_id` column type (migration `003`), **restart the API**
> so pgx drops prepared statements that still assume the old type.

## JWT / auth

```bash
make gentoken                                   # uses JWT_SECRET, user dev-user
curl -X DELETE localhost:8080/api/v1/links/CODE \
  -H "Authorization: Bearer <token>"
```

Auth is required for `DELETE /api/v1/links/{short_code}` and optional on
`POST /api/v1/shorten` (present ⇒ the link is owned by the token's `user_id`).
Anonymous links cannot be deleted through the API.

## Verifying a deploy

```bash
make ready && make e2e
```

`make e2e` checks health, readiness, shorten, cache miss→hit metric movement,
JWT-gated delete, validation errors, and the metrics endpoint.

## Failure modes and response

| Failure | Symptom | Expected behaviour | Response |
|---------|---------|--------------------|----------|
| Redis down | `cache_misses_total` climbs, latency up | Fail-open: redirects hit Postgres, limits not enforced (`NoopCache`) | Restore Redis; check memory/eviction; alert was expected |
| Postgres primary down | `/ready` 503, redirects 5xx | `GET /ready` reports the DB error | Failover to replica / HA; check connections |
| Kafka down (Phase 4) | — | Redirect still 302; click loss acceptable short-term | Monitor producer errors; buffer/retry |
| Hot shard (post-shard) | one shard saturated | elevated p99 | Cache viral keys; rebalance (ADR [0002](./adr/0002-hash-sharding-vs-range-sharding.md)) |
| Snowflake clock skew | duplicate/rare ID errors | generator waits for the clock; NTP expected | Check NTP; investigate host clock |

See ADR [0005](./adr/0005-fail-open-vs-fail-closed-on-redis-outage.md) for the
fail-open rationale.

## Rollback

1. `make docker-down`
2. Check out the last good commit / image tag.
3. If a migration was applied, run `make migrate-down` to the matching version.
4. `make docker-up && make e2e`.
