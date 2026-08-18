#!/usr/bin/env sh
set -eu

if command -v go >/dev/null 2>&1; then
  go test ./...
  exit 0
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: Go binary not found, and docker is unavailable."
  echo "Instalasi go atau Docker diperlukan untuk menjalankan test suite."
  exit 1
fi

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
docker run --rm \
  -v "$PROJECT_ROOT:/src" \
  -w /src \
  golang:1.22 \
  sh -lc "go test ./..."
