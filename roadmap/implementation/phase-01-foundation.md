# Phase 1 Implementation Log — Orientation & Foundation

**Roadmap:** `roadmap/phases/01-orientation-and-foundation.md`  
**Git commit:** `aab5677` — *Implement Phase 1 foundation with colored dev logging.*  
**Date:** 2026-06-14

---

## Goal

Bootstrapped the project so the API starts, Postgres runs in Docker, migrations apply, and a health check proves the HTTP stack works—before any shorten/redirect feature code.

---

## What was delivered

| Deliverable | Status |
|-------------|--------|
| Go module + `cmd/api` entrypoint | Done |
| Config from env / `.env` | Done |
| Structured logging (colored console in dev) | Done |
| Docker Compose (Postgres + migrate + API) | Done |
| Migration `001_url_mappings` | Done |
| `GET /health` | Done |
| Makefile + `.env.example` | Done |

---

## Project layout created

```
NexusLink/
├── cmd/api/main.go              # Application entrypoint
├── internal/
│   ├── config/                  # Env-based configuration
│   ├── logging/                 # slog handlers (colored + plain)
│   └── server/                  # HTTP server + /health
├── deploy/
│   ├── Dockerfile               # Multi-stage Go build
│   └── docker-compose.yml       # postgres + migrate + api
├── migrations/
│   └── 001_url_mappings.*.sql
├── scripts/migrate.sh
├── Makefile
├── .env.example
└── .gitignore
```

---

## File-by-file changes

### `cmd/api/main.go`

**Purpose:** Wire config → logger → HTTP server → graceful shutdown.

**Flow at Phase 1:**

1. `config.Load()` reads `PORT`, `ENV`, `LOG_LEVEL`, `DATABASE_URL`, etc.
2. `logging.Setup()` returns an `*slog.Logger` writing to stdout (colored in dev) and optionally a log file.
3. `server.New()` registers only `GET /health`.
4. Signal handler (`SIGINT`/`SIGTERM`) calls `srv.Shutdown` with a 10s timeout.

**Note:** Phase 2 later added Postgres pool + shorten/redirect. Phase 3 replaced wiring with Redis, JWT, metrics, etc. The skeleton (config, logging, graceful shutdown) remains.

---

### `internal/config/config.go`

**Struct (Phase 1 fields):**

| Field | Env var | Default |
|-------|---------|---------|
| `Env` | `ENV` | `development` |
| `Port` | `PORT` | `8080` |
| `LogLevel` | `LOG_LEVEL` | `info` |
| `LogDir` | `LOG_DIR` | `logs/api` |
| `LogFile` | `LOG_FILE` | `app.log` |
| `DatabaseURL` | `DATABASE_URL` | `postgres://nexuslink:nexuslink@localhost:5432/nexuslink?sslmode=disable` |

**Behavior:**

- Uses `github.com/joho/godotenv` to load `.env` if present.
- `Load()` returns an error on invalid `PORT`.
- `Addr()` returns `:8080` style listen address.

**Tests:** `internal/config/config_test.go` — defaults, env override, invalid port.

---

### `internal/logging/`

**Packages/files:**

| File | Role |
|------|------|
| `logging.go` | `Setup()`, `New()`, level parsing |
| `handler.go` | Custom `slog.Handler` implementations |
| `colors.go` | ANSI colors for dev console |
| `handler_test.go` | Handler output tests |

**Design:**

- **Development:** colored, human-readable logs on stdout.
- **Production:** JSON logs on stdout (structured for log aggregators).
- Phase 2 added **file logging** (`file.go`, `multi_handler.go`) — see Phase 2 doc.

**HTTP request logging:** `LogHTTPRequest()` called from `server.requestLogger` middleware — logs method, path, status, duration per request.

---

### `internal/server/server.go`

**Phase 1 responsibilities:**

- `http.ServeMux` with Go 1.22+ method-aware routes: `GET /health`.
- `requestLogger` middleware wraps all routes.
- `ReadHeaderTimeout: 5s` on `http.Server` (slowloris mitigation).
- JSON helper `writeJSON`.

**`GET /health` response:**

```json
{"status":"ok"}
```

Always `200`. Used for **liveness** (process is up). Phase 3 added `GET /ready` for **readiness** (dependencies healthy).

**Tests:** `internal/server/server_test.go` — health endpoint returns 200 and `status: ok`.

---

### `deploy/docker-compose.yml` (Phase 1 version)

**Services:**

| Service | Image | Role |
|---------|-------|------|
| `postgres` | `postgres:16-alpine` | Primary database |
| `migrate` | `migrate/migrate:v4.17.1` | Runs SQL migrations once, exits |
| `api` | built from `deploy/Dockerfile` | Go API binary |

**Startup order:** `postgres` (healthcheck) → `migrate` (success) → `api`.

**Postgres defaults:** user/password/db = `nexuslink`.

---

### `deploy/Dockerfile`

Multi-stage build:

1. **Builder:** `golang:1.22-alpine`, `GOTOOLCHAIN=auto` (allows `go.mod` to request newer toolchain).
2. **Runtime:** `alpine:3.20`, copies static binary `/app/api`, exposes `8080`.

---

### `migrations/001_url_mappings.up.sql`

Creates the core table used by all later phases:

```sql
CREATE TABLE url_mappings (
    short_code   VARCHAR(16) PRIMARY KEY,
    long_url     TEXT NOT NULL,
    snowflake_id BIGINT UNIQUE,
    user_id      UUID,                    -- changed to TEXT in migration 003
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    http_status  SMALLINT NOT NULL DEFAULT 302
);
```

**Indexes:** `user_id`, `created_at`.

**Design intent:** `short_code` is the public key; `snowflake_id` stores the internal numeric ID for generated links (NULL for custom aliases); `is_active` supports soft delete later.

---

### `scripts/migrate.sh`

Runs `migrate` CLI locally if installed, otherwise spins a one-off Docker container with `--network host`.

```bash
./scripts/migrate.sh up    # apply migrations
./scripts/migrate.sh down  # rollback one step
```

---

### `Makefile` (Phase 1 targets)

| Target | Action |
|--------|--------|
| `run` | `go run ./cmd/api` |
| `build` | `go build -o bin/api ./cmd/api` |
| `test` | `go test ./...` |
| `docker-up` | `docker compose -f deploy/docker-compose.yml up -d --build` |
| `docker-down` | tear down stack |
| `migrate-up` / `migrate-down` | wrapper around `scripts/migrate.sh` |
| `health` | `curl` `GET /health` |

---

### Supporting files

| File | Purpose |
|------|---------|
| `.env.example` | Documents env vars for local dev |
| `.gitignore` | Ignores `bin/`, `.env`, logs, etc. |
| `.dockerignore` | Keeps Docker build context small |
| `go.mod` | Module `github.com/JathinShyam/NexusLink` |
| `README.md` | Updated with quick-start instructions |

---

## Request flow (Phase 1 only)

```
Browser/curl
    → GET /health
    → server.requestLogger
    → handleHealth
    → JSON {"status":"ok"}
```

No database connection was required for `/health` in Phase 1. The app could start even if Postgres was down (DB wiring came in Phase 2).

---

## Dependencies added (Phase 1)

```
github.com/joho/godotenv v1.5.1
```

(`pgx` and others added in Phase 2.)

---

## How to verify Phase 1

```bash
cp .env.example .env
make docker-up
make health
# Expect: {"status":"ok"}

make migrate-up   # if running API locally against Docker Postgres
make run          # optional: run API on host
```

**Exit criteria met:**

- App starts locally via Docker or `make run`.
- Postgres runs via Compose.
- Migration applies without error.
- Health returns 200.
- You can explain: request → mux → handler → JSON response.

---

## What Phase 1 intentionally did NOT include

- Shorten or redirect endpoints
- Redis, Kafka, auth, metrics
- `internal/storage` implementation (only `.gitkeep` placeholder)
- Unit tests beyond config, logging, server health

These are all Phase 2+ concerns.
