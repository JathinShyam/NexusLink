#!/usr/bin/env bash
# Tail service log files with colors preserved (use -R for ANSI).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE="${1:-api}"
LINES="${2:-50}"

case "$SERVICE" in
  api)
    FILE="$ROOT_DIR/logs/api/app.log"
    ;;
  postgres)
    FILE="$ROOT_DIR/logs/postgres/postgres.log"
    ;;
  migrate)
    FILE="$ROOT_DIR/logs/migrate/migrate.log"
    ;;
  all)
    tail -n "$LINES" -F \
      "$ROOT_DIR/logs/api/app.log" \
      "$ROOT_DIR/logs/postgres/postgres.log" \
      "$ROOT_DIR/logs/migrate/migrate.log" 2>/dev/null || true
    exit 0
    ;;
  *)
    echo "Usage: $0 [api|postgres|migrate|all] [lines]" >&2
    exit 1
    ;;
esac

if [[ ! -f "$FILE" ]]; then
  echo "Log file not found: $FILE" >&2
  echo "Start the stack first: make docker-up" >&2
  exit 1
fi

if [[ "${FOLLOW:-1}" == "1" ]]; then
  tail -n "$LINES" -f "$FILE"
else
  tail -n "$LINES" "$FILE"
fi
