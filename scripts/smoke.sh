#!/usr/bin/env bash
set -euo pipefail
BASE="${1:-http://127.0.0.1:8080/api}"
echo "health: $(curl -sf "$BASE/health")"
USER="smoke_$RANDOM"
REG=$(curl -sf -X POST "$BASE/auth/register" -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"email\":\"$USER@example.com\",\"password\":\"secret12\"}")
TOKEN=$(python3 -c "import json,sys; print(json.load(sys.stdin)['token'])" <<< "$REG")
AUTH="Authorization: Bearer $TOKEN"
SESS=$(curl -sf -X POST "$BASE/sessions/start" -H "$AUTH" -H 'X-Timezone: Asia/Shanghai')
SID=$(python3 -c "import json,sys; print(json.load(sys.stdin)['id'])" <<< "$SESS")
KEYS=$(python3 -c "import json,sys; print(' '.join(x['key'] for x in json.load(sys.stdin)['steps']))" <<< "$SESS")
for k in $KEYS; do
  curl -sf -X PATCH "$BASE/sessions/$SID/steps/$k" -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"status":"completed","actual_duration_seconds":10}' >/dev/null
done
curl -sf -X POST "$BASE/sessions/$SID/complete" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"mood":5,"note":"smoke"}' >/dev/null
echo "SMOKE_OK user=$USER session=$SID"
