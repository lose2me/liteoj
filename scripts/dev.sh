#!/usr/bin/env bash
# Start backend + frontend + (reminder) go-judge for local development.
# go-judge is expected to already be running at $JUDGE_BASE_URL — start it yourself.
set -e
cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  cp .env.example .env
  echo "[dev] copied .env.example to .env — please review values (JWT_SECRET, ADMIN_INIT_PASSWORD, ...)"
fi

echo "[dev] reminder: start go-judge separately, e.g.  ./go-judge -http-addr :5050"
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
