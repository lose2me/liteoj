#!/usr/bin/env bash
# Produce a single static binary with the embedded frontend.
set -e
cd "$(dirname "$0")/.."

echo "[build] pnpm build"
(cd frontend && pnpm install --prod=false && pnpm build)

echo "[build] copy dist into backend/internal/web/dist"
rm -rf backend/internal/web/dist
mkdir -p backend/internal/web/dist
cp -r frontend/dist/. backend/internal/web/dist/

echo "[build] go build"
(cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../liteoj ./cmd/liteoj)

echo "[build] done: ./liteoj"
