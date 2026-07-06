#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_SECRET="${JWT_SECRET:-dev-secret-change-me}"
USER_ID="${USER_ID:-dev-user}"
LONG_URL="${LONG_URL:-https://example.com/docs/nexuslink-phase-3}"
ALIAS="p3$(date +%s)"

echo "==> Ready check"
curl -sf "$BASE_URL/ready" | sed 's/^/    /'
echo

echo "==> Issue dev JWT for user: $USER_ID"
TOKEN="$(JWT_SECRET="$JWT_SECRET" go run ./cmd/gentoken -user "$USER_ID")"
echo "    token issued (${#TOKEN} chars)"

echo "==> Create owned short link"
CREATE_RESPONSE="$(curl -sf -X POST "$BASE_URL/api/v1/shorten" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"long_url\":\"$LONG_URL\",\"custom_alias\":\"$ALIAS\"}")"
echo "$CREATE_RESPONSE" | sed 's/^/    /'

SHORT_URL="$(echo "$CREATE_RESPONSE" | sed -n 's/.*"short_url":"\([^"]*\)".*/\1/p')"

echo "==> First redirect (cache miss expected)"
curl -sI "$SHORT_URL" | sed 's/^/    /'

echo "==> Second redirect (cache hit expected)"
curl -sI "$SHORT_URL" | sed 's/^/    /'

echo "==> Metrics snapshot (cache counters)"
curl -sf "$BASE_URL/metrics" | grep -E 'cache_(hits|misses)_total' | sed 's/^/    /' || true

echo "==> Delete owned link"
curl -sf -X DELETE "$BASE_URL/api/v1/links/$ALIAS" \
  -H "Authorization: Bearer $TOKEN" \
  -w "    HTTP %{http_code}\n" -o /dev/null

echo "==> Redirect after delete (expect 404)"
curl -sI "$SHORT_URL" | sed 's/^/    /'

echo "==> Phase 3 demo complete"
