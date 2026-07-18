# NexusLink web UI

React + Vite frontend for the NexusLink API.

## Run

**Docker:**

```bash
make docker-up
# open http://localhost:3000
```

**Vite (from the repo root, API must be up):**

```bash
make docker-up
make web-install
make web-dev
```

Open [http://localhost:5173](http://localhost:5173).

## Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Dev server with proxy to `:8080` |
| `npm run build` | Typecheck + production build |
| `npm run preview` | Preview production build |

## Env

See [`.env.example`](./.env.example). Leave `VITE_API_BASE_URL` empty to use the Vite proxy.
