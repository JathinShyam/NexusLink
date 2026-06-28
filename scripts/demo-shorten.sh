#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
LONG_URL="${LONG_URL:-https://example.com/docs/nexuslink-phase-2}"

echo "==> Creating short link for: $LONG_URL"
CREATE_RESPONSE="$(curl -sf -X POST "$BASE_URL/api/v1/shorten" \
  -H 'Content-Type: application/json' \
  -d "{\"long_url\":\"$LONG_URL\"}")"

echo "$CREATE_RESPONSE" | sed 's/^/    /'

SHORT_CODE="$(echo "$CREATE_RESPONSE" | sed -n 's/.*"short_code":"\([^"]*\)".*/\1/p')"
SHORT_URL="$(echo "$CREATE_RESPONSE" | sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p')"

if [[ -z "$SHORT_CODE" ]]; then
  echo "Failed to parse short_code from response" >&2
  exit 1
fi

echo "==> Redirect check (expect HTTP 302)"
curl -sI "$SHORT_URL" | sed 's/^/    /'

echo "==> Done. Open this in a browser: $SHORT_URL"
