#!/usr/bin/env bash
# Produce a single static binary plus the root ./dist frontend bundle.
set -e
cd "$(dirname "$0")/.."

echo "[build] pnpm build"
(cd frontend && pnpm install --prod=false && pnpm build)

echo "[build] copy dist into ./dist"
rm -rf dist
mkdir -p dist
cp -r frontend/dist/. dist/

echo "[build] go build"
(cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../liteoj ./cmd/liteoj)

echo "[build] done: ./liteoj"
