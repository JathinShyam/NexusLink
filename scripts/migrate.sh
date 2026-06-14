#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

DATABASE_URL="${DATABASE_URL:-postgres://nexuslink:nexuslink@localhost:5432/nexuslink?sslmode=disable}"

run_migrate() {
  if command -v migrate >/dev/null 2>&1; then
    migrate -path migrations -database "$DATABASE_URL" "$@"
    return
  fi

  docker run --rm \
    -v "$ROOT_DIR/migrations:/migrations" \
    --network host \
    migrate/migrate:v4.17.1 \
    -path=/migrations \
    -database "$DATABASE_URL" \
    "$@"
}

case "${1:-up}" in
  up)
    run_migrate up
    ;;
  down)
    run_migrate down 1
    ;;
  *)
    echo "Usage: $0 [up|down]" >&2
    exit 1
    ;;
esac
