#!/usr/bin/env sh
set -eu

BASE_URL="${BASE_URL:-https://be-sharing-vision.meetsin.id}"
FE_URL="${FE_URL:-https://sharing-vision-frontend-two.vercel.app}"

printf "[1/6] GET %s/\n" "$BASE_URL"
curl -ksS -m 20 "$BASE_URL/" || exit 1

printf "[2/6] GET %s/health\n" "$BASE_URL"
curl -ksS -m 20 "$BASE_URL/health" || exit 1

printf "[3/6] GET %s/ready\n" "$BASE_URL"
curl -ksS -m 20 "$BASE_URL/ready" || exit 1

printf "[4/6] GET %s/article/10/0\n" "$BASE_URL"
curl -ksS -m 20 "$BASE_URL/article/10/0" || exit 1

printf "[5/6] FE proxy check: %s/api/article/10/0\n" "$FE_URL"
curl -ksS -m 20 "$FE_URL/api/article/10/0" || exit 1

printf "[6/6] FE dashboard page\n"
curl -ksS -m 20 "$FE_URL/?api=$BASE_URL" | head -n 20 || exit 1

printf "OK: live verification finished.\n"
