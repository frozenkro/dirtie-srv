#!/usr/bin/env bash
set -euo pipefail

# Smoke tests for dirtie-srv endpoints on the k3s deployment
# Usage: ./smoke-test.sh [BASE_URL]
# Default target is rpic1 on Tailscale

BASE_URL="${1:-http://rpic1:8080}"
COOKIE_JAR="$(mktemp /tmp/dirtie-smoke-cookies.XXXXXX)"

cleanup() { rm -f "$COOKIE_JAR"; }
trap cleanup EXIT

err() { echo "[FAIL] $*" >&2; exit 1; }
ok()  { echo "[PASS] $*"; }

ACTION() {
  printf "\n--- %s ---\n" "$*"
}

# Unauthenticated endpoints
ACTION "GET ${BASE_URL}/"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/")
[[ "$CODE" == "200" ]] || err "GET / expected 200, got ${CODE}"
ok "GET / => ${CODE}"

ACTION "POST ${BASE_URL}/users"
USER_EMAIL="smoke-$(date +%s)@example.com"
CREATE_BODY='{"email":"'"${USER_EMAIL}"'","password":"SmokeP@ss123","name":"Smoke Test"}'
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$CREATE_BODY" "${BASE_URL}/users")
[[ "$CODE" == "200" ]] || err "POST /users expected 200, got ${CODE}"
ok "POST /users (create user) => ${CODE}"

ACTION "POST ${BASE_URL}/login"
LOGIN_BODY='{"email":"'"${USER_EMAIL}"'","password":"SmokeP@ss123"}'
CODE=$(curl -s -c "$COOKIE_JAR" -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$LOGIN_BODY" "${BASE_URL}/login")
[[ "$CODE" == "200" ]] || err "POST /login expected 200, got ${CODE}"
ok "POST /login => ${CODE}"

# Verify cookie exists
if ! grep -q "dirtie.auth" "$COOKIE_JAR"; then
  err "No dirtie.auth cookie received after login"
fi
ok "Cookie captured"

# Authenticated endpoints
ACTION "GET ${BASE_URL}/devices"
CODE=$(curl -s -b "$COOKIE_JAR" -o /dev/null -w "%{http_code}" "${BASE_URL}/devices")
# 200 if OK, 500/404 if no devices or DB issue — anything but 401/403 is usually OK for a smoke test
if [[ "$CODE" == "401" || "$CODE" == "403" ]]; then
  err "GET /devices expected authed, got ${CODE}"
fi
ok "GET /devices => ${CODE}"

ACTION "POST ${BASE_URL}/devices/createProvision?displayName=SmokeDevice"
CODE=$(curl -s -b "$COOKIE_JAR" -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/devices/createProvision?displayName=SmokeDevice")
if [[ "$CODE" == "401" || "$CODE" == "403" ]]; then
  err "POST /devices/createProvision expected authed, got ${CODE}"
fi
ok "POST /devices/createProvision => ${CODE}"

ACTION "POST ${BASE_URL}/logout"
CODE=$(curl -s -b "$COOKIE_JAR" -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/logout")
[[ "$CODE" == "200" ]] || err "POST /logout expected 200, got ${CODE}"
ok "POST /logout => ${CODE}"

# After logout, authed request should fail
ACTION "GET ${BASE_URL}/devices (after logout)"
CODE=$(curl -s -b "$COOKIE_JAR" -o /dev/null -w "%{http_code}" "${BASE_URL}/devices")
[[ "$CODE" == "401" || "$CODE" == "403" ]] || err "GET /devices after logout expected 401/403, got ${CODE}"
ok "GET /devices after logout => ${CODE}"

ACTION "POST ${BASE_URL}/pw/reset?email=${USER_EMAIL}"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/pw/reset?email=${USER_EMAIL}")
[[ "$CODE" == "200" ]] || err "POST /pw/reset expected 200, got ${CODE}"
ok "POST /pw/reset => ${CODE}"

printf "\n=== All smoke tests passed ===\n"
