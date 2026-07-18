# Frontend UI — Signal Board Landing

**Status:** Complete (pre–Phase 4)  
**Location:** [`web/`](../../web/)

## Goal

Ship a brand-first React UI for the existing Phase 3 API: shorten links, optional JWT ownership, and soft-delete by short code.

## Stack

- React 19 + Vite + TypeScript under `web/`
- Fonts: Syne (display) + DM Sans (body)
- Visual: “signal board” — slate field, teal accent, full-bleed path graphic in the hero

## What it calls

| UI action | API |
|-----------|-----|
| Shorten | `POST /api/v1/shorten` (optional `Authorization: Bearer`) |
| Delete | `DELETE /api/v1/links/{short_code}` (JWT required) |

No list-links UI — backend has no list endpoint yet.

## Local run

**Docker (full stack including UI):**

```bash
make docker-up
# UI:  http://localhost:3000
# API: http://localhost:8080
```

Nginx serves the built SPA and reverse-proxies `/api`, `/health`, and `/ready` to the API service.

**Vite (hot reload during UI work):**

```bash
make docker-up    # API + Postgres + Redis
make web-install  # once
make web-dev      # http://localhost:5173
```

Dev proxy (default): Vite forwards `/api`, `/health`, `/ready` to `:8080`.

Direct API calls: set `VITE_API_BASE_URL=http://localhost:8080` and ensure `CORS_ORIGINS` includes the Vite origin.

JWT for owned links:

```bash
make gentoken
# paste token into Manage with JWT panel
```

## Backend support

- CORS middleware: `internal/cors`
- Config: `CORS_ORIGINS` (default includes `localhost:5173` and `127.0.0.1:5173`)
- Wired in `internal/server` + `cmd/api/main.go`

## Key files

```
web/src/components/Hero.tsx
web/src/components/ShortenForm.tsx
web/src/components/ResultPanel.tsx
web/src/components/ManagePanel.tsx
web/src/api/client.ts
internal/cors/cors.go
```

## Makefile

| Target | Action |
|--------|--------|
| `web-install` | `npm install` in `web/` |
| `web-dev` | Vite dev server |
| `web-build` | production build to `web/dist` |
