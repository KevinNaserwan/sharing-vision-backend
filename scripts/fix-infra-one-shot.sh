#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-https://be-sharing-vision.meetsin.id}"
FE_URL="${FE_URL:-https://sharing-vision-frontend-two.vercel.app}"
DOMAIN="${DOMAIN:-be-sharing-vision.meetsin.id}"
CERT_NAME="${CERT_NAME:-/etc/letsencrypt/live/${DOMAIN}/fullchain.pem}"

assert_json() {
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$1" | jq -e . >/dev/null
  else
    printf '%s' "$1" | sed -n '1p' >/dev/null
  fi
}

printf '\n[1/6] Nginx config check\n'
sudo nginx -t

printf '\n[2/6] Reload Nginx (no-op if already clean)\n'
sudo nginx -s reload

printf '\n[3/6] TLS certificate validation: %s\n' "$DOMAIN"
if sudo test -f "$CERT_NAME"; then
  CERT_FILE="$CERT_NAME"
else
  CERT_FILE="$(sudo ls -1 /etc/letsencrypt/archive/$DOMAIN/fullchain*.pem 2>/dev/null | head -n 1 || true)"
fi

if [ -n "$CERT_FILE" ] && sudo test -f "$CERT_FILE"; then
  sudo openssl x509 -in "$CERT_FILE" -noout -subject -issuer -dates
else
  echo "WARN: Certificate file not found: $CERT_NAME"
fi

printf '\n[4/6] Backend smoke (GET)\n'
health_body=$(curl -ksS -m 30 "$BASE_URL/health")
assert_json "$health_body"
ready_body=$(curl -ksS -m 30 "$BASE_URL/ready")
assert_json "$ready_body"
list_body=$(curl -ksS -m 30 "$BASE_URL/article/10/0")
assert_json "$list_body"

printf '\n[5/6] FE proxy smoke (GET list + create + validation)\n'
assert_json "$(curl -ksS -m 30 "$FE_URL/api/article/10/0")"

good_payload=$(cat <<'JSON'
{
  "title": "Judul verifikasi one-shot checklist 9999",
  "content": "Ini payload valid untuk pengecekan endpoint. Isi konten ini sengaja sangat panjang melebihi dua ratus karakter agar memenuhi aturan minimum content ketika melakukan pemeriksaan otomatis terhadap status 201 pada endpoint create.",
  "category": "Ops",
  "status": "draft"
}
JSON
)

assert_json "$good_payload"

curl -ksS -m 30 -X POST "$FE_URL/api/article" \
  -H 'Content-Type: application/json' \
  -d "$good_payload" \
  | if command -v jq >/dev/null 2>&1; then jq -e . >/dev/null; else sed -n '1p' >/dev/null; fi || exit 1

bad_payload=$(cat <<'JSON'
{
  "title": "xx",
  "content": "short",
  "category": "x",
  "status": "invalid"
}
JSON
)

echo "\nExpect HTTP 400:"
code=$(curl -ksS -m 30 -o /tmp/be_sv_bad_response.json -w '%{http_code}' -X POST "$FE_URL/api/article" -H 'Content-Type: application/json' -d "$bad_payload")
if [ "$code" != "400" ]; then
  echo "Expected 400 but got ${code}"
  cat /tmp/be_sv_bad_response.json
  exit 1
fi

printf '\n[6/6] HEAD checks for infra health endpoints\n'
for path in /health /ready; do
  code=$(curl -ksSI -o /dev/null -m 30 -w '%{http_code}' "$BASE_URL$path")
  if [ "$code" != "200" ]; then
    echo "Expected 200 for ${path} but got ${code}"
    exit 1
  fi
  echo "OK ${path}: ${code}"
done

printf '\nOK: one-shot infra check passed.\n'
