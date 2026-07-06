#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
JWT_SECRET="${JWT_SECRET:-dev-secret-change-me}"
USER_ID="${USER_ID:-dev-user}"
FAILURES=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; FAILURES=$((FAILURES + 1)); }

assert_status() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then
    pass "$label (HTTP $actual)"
  else
    fail "$label (expected HTTP $expected, got $actual)"
  fi
}

echo "==> E2E: health"
assert_status "GET /health" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/health")"

echo "==> E2E: ready"
assert_status "GET /ready" 200 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/ready")"

echo "==> E2E: shorten"
CREATE_RESPONSE="$(curl -sf -X POST "$BASE_URL/api/v1/shorten" \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com/e2e-test"}')"
SHORT_CODE="$(echo "$CREATE_RESPONSE" | sed -n 's/.*"short_code":"\([^"]*\)".*/\1/p')"
if [[ -n "$SHORT_CODE" ]]; then
  pass "POST /api/v1/shorten returned short_code=$SHORT_CODE"
else
  fail "POST /api/v1/shorten missing short_code"
fi

metrics_body() { curl -sf "$BASE_URL/metrics"; }

metric_value() {
  local name="$1"
  metrics_body | awk -v name="$name" '$1 == name { print $2; found=1 } END { if (!found) print "0" }'
}

echo "==> E2E: redirect + cache"
# Shorten warms cache on create; evict key so first redirect is a miss.
redis-cli -u "${REDIS_URL:-redis://localhost:6379}" DEL "url:$SHORT_CODE" >/dev/null 2>&1 || true

MISS_BEFORE="$(metric_value cache_misses_total)"
HITS_BEFORE="$(metric_value cache_hits_total)"

assert_status "GET /$SHORT_CODE first (miss)" 302 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/$SHORT_CODE")"
assert_status "GET /$SHORT_CODE second (hit)" 302 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/$SHORT_CODE")"

MISS_AFTER="$(metric_value cache_misses_total)"
HITS_AFTER="$(metric_value cache_hits_total)"

if awk "BEGIN {exit !($MISS_AFTER > $MISS_BEFORE)}"; then
  pass "cache_misses_total increased ($MISS_BEFORE -> $MISS_AFTER)"
else
  fail "cache_misses_total did not increase ($MISS_BEFORE -> $MISS_AFTER)"
fi

if awk "BEGIN {exit !($HITS_AFTER > $HITS_BEFORE)}"; then
  pass "cache_hits_total increased ($HITS_BEFORE -> $HITS_AFTER)"
else
  fail "cache_hits_total did not increase ($HITS_BEFORE -> $HITS_AFTER)"
fi

echo "==> E2E: auth + delete"
TOKEN="$(JWT_SECRET="$JWT_SECRET" go run ./cmd/gentoken -user "$USER_ID")"
OWNED_ALIAS="e2e$(date +%s)"
AUTH_CREATE="$(curl -sf -X POST "$BASE_URL/api/v1/shorten" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"long_url\":\"https://example.com/e2e-owned\",\"custom_alias\":\"$OWNED_ALIAS\"}")"
if echo "$AUTH_CREATE" | grep -q "$OWNED_ALIAS"; then
  pass "authenticated shorten created $OWNED_ALIAS"
else
  fail "authenticated shorten failed: $AUTH_CREATE"
fi

assert_status "DELETE without auth" 401 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/links/$OWNED_ALIAS")"
assert_status "DELETE with auth" 204 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/links/$OWNED_ALIAS" -H "Authorization: Bearer $TOKEN")"
assert_status "GET deleted link" 404 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/$OWNED_ALIAS")"

echo "==> E2E: validation errors"
assert_status "invalid url" 400 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/shorten" -H 'Content-Type: application/json' -d '{"long_url":"ftp://bad"}')"
assert_status "not found redirect" 404 "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/does-not-exist-e2e")"

echo "==> E2E: metrics endpoint"
METRICS_CODE="$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/metrics")"
assert_status "GET /metrics" 200 "$METRICS_CODE"
if curl -sf "$BASE_URL/metrics" | grep -q 'http_requests_total'; then
  pass "metrics include http_requests_total"
else
  fail "metrics missing http_requests_total"
fi

echo
if [[ "$FAILURES" -eq 0 ]]; then
  echo "All E2E checks passed."
  exit 0
fi

echo "$FAILURES check(s) failed."
exit 1
