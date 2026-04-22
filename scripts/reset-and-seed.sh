#!/usr/bin/env bash
# scripts/reset-and-seed.sh —— 一键重置 LiteOJ 开发数据库：
#   1. 停掉正在运行的 liteoj.exe（按 PID 文件 / 端口 fallback）
#   2. 备份 ./data/liteoj.db（包括 WAL / SHM）到 ./data/backups/
#   3. 删除 DB 文件，重启后端，让 seed 重建
#
# 注意：**不动** go-judge —— 沙箱进程由 C:\WSL\*.bat 管，脚本只负责 OJ 侧。
# 运行前请确认已在项目根目录，且 ./liteoj.exe 已编译。

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DATA_DIR="$ROOT/data"
DB_FILE="$DATA_DIR/liteoj.db"
BACKUPS="$DATA_DIR/backups"
PID_FILE="/tmp/liteoj.pid"
PORT="${LITEOJ_PORT:-8080}"

mkdir -p "$BACKUPS"

echo "[reset] project root: $ROOT"

# 1) 停后端
stop_backend() {
  if [[ -f "$PID_FILE" ]]; then
    local pid
    pid="$(cat "$PID_FILE")"
    if kill -0 "$pid" 2>/dev/null; then
      echo "[reset] killing liteoj pid=$pid (from pid file)"
      kill "$pid" 2>/dev/null || true
      sleep 1
    fi
    rm -f "$PID_FILE"
  fi
  # fallback：按端口杀（Windows 本地 bash 可能没 fuser，先 lsof 再 fuser）
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -ti tcp:$PORT 2>/dev/null || true)"
    if [[ -n "$pids" ]]; then
      echo "[reset] killing processes on :$PORT → $pids"
      echo "$pids" | xargs -r kill 2>/dev/null || true
      sleep 1
    fi
  fi
}

stop_backend

# 2) 备份
if [[ -f "$DB_FILE" ]]; then
  ts="$(date +%Y%m%d-%H%M%S)"
  dest="$BACKUPS/liteoj.db.$ts"
  cp "$DB_FILE" "$dest"
  [[ -f "$DB_FILE-wal" ]] && cp "$DB_FILE-wal" "$dest-wal" || true
  [[ -f "$DB_FILE-shm" ]] && cp "$DB_FILE-shm" "$dest-shm" || true
  echo "[reset] backed up → $dest"
fi

# 3) 删 DB
rm -f "$DB_FILE" "$DB_FILE-wal" "$DB_FILE-shm"
echo "[reset] removed $DB_FILE"

# 4) 重启后端（前台；Ctrl+C 可退出）
if [[ ! -x "$ROOT/liteoj.exe" && ! -x "$ROOT/liteoj" ]]; then
  echo "[reset] WARN: liteoj binary not found; run 'cd backend && go build -o ../liteoj.exe ./cmd/liteoj' first"
  exit 0
fi
BIN="$ROOT/liteoj.exe"
[[ -x "$BIN" ]] || BIN="$ROOT/liteoj"

echo "[reset] starting $BIN ..."
"$BIN" &
echo $! > "$PID_FILE"
wait $!
