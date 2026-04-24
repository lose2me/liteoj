#!/usr/bin/env bash
# Start backend + frontend for local development.
# go-judge is expected to already be running; start it yourself first.
set -e
cd "$(dirname "$0")/.."

if [ ! -f config.toml ]; then
  cp config.example.toml config.toml
  echo "[dev] copied config.example.toml to config.toml — please review judge/jwt/ai values"
fi

echo "[dev] reminder: start go-judge separately first"
echo "[dev]   Windows host: C:\\WSL\\start-go-judge.bat"
echo "[dev]   Liveness check: curl http://127.0.0.1:5050/version"
echo

# backend in background
(
  cd backend
  go run ./cmd/liteoj &
  echo $! > /tmp/liteoj-backend.pid
) &

# frontend (foreground)
cd frontend
pnpm install
pnpm dev
