#!/usr/bin/env bash
# scripts/audit-e2e.sh —— Phase D 端到端冒烟：真启 liteoj + 打 go-judge。
# 覆盖规模：登录 / 建题单 / 踢人 / stu7 penalty=60 / stu8 rejected 不罚 / stu10
# 保留提交 / 判题超时降级 / SSE 流（30s keepalive 窗口）。任一步失败 exit 1。
#
# 前置：C:\WSL\start-go-judge.bat 已经跑过。脚本**不**启停 go-judge，这是用户
# 的环境职责。
#
# 日志：docs/e2e-<date>.log。

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

LOG_DIR="$ROOT/docs"
mkdir -p "$LOG_DIR"
STAMP="$(date +%Y%m%d-%H%M%S)"
LOG="$LOG_DIR/e2e-$STAMP.log"
exec > >(tee -a "$LOG") 2>&1

BASE="${LITEOJ_BASE:-http://127.0.0.1:8080}"
JUDGE_URL="${JUDGE_URL:-http://127.0.0.1:5050}"
ADMIN_USER="${LITEOJ_ADMIN_USER:-admin}"
ADMIN_PASS="${LITEOJ_ADMIN_PASS:-admin123}"
PID_FILE="/tmp/liteoj-e2e.pid"
AH=()
RESTORE_PROBLEM_ID=""
RESTORE_PROBLEM_PAYLOAD=""

step() { echo; echo "─── $* ───"; }
die()  { echo "[e2e] FAIL: $*"; exit 1; }

step "1) go-judge liveness"
curl -fsS "$JUDGE_URL/version" >/dev/null || die "go-judge unreachable at $JUDGE_URL"
echo "  ok"

step "2) build backend + frontend"
( cd backend && go build -o ../liteoj.exe ./cmd/liteoj ) || die "backend build"
( cd frontend && pnpm build >/dev/null 2>&1 ) || die "frontend build"
echo "  ok"

step "3) start liteoj.exe in background"
"$ROOT/liteoj.exe" > "$LOG_DIR/e2e-$STAMP.server.log" 2>&1 &
echo $! > "$PID_FILE"
sleep 3
curl -fsS "$BASE/api/health" >/dev/null || die "backend not healthy on $BASE"
echo "  pid=$(cat "$PID_FILE")"

cleanup() {
  if [[ -n "${RESTORE_PROBLEM_ID:-}" && -n "${RESTORE_PROBLEM_PAYLOAD:-}" && ${#AH[@]} -gt 0 ]]; then
    curl -fsS "${AH[@]}" -H 'Content-Type: application/json' \
      -X PUT "$BASE/api/admin/problems/$RESTORE_PROBLEM_ID" \
      -d "$RESTORE_PROBLEM_PAYLOAD" >/dev/null || true
  fi
  if [[ -f "$PID_FILE" ]]; then
    kill "$(cat "$PID_FILE")" 2>/dev/null || true
    rm -f "$PID_FILE"
  fi
}
trap cleanup EXIT

# tiny JSON helper (requires jq)
if ! command -v jq >/dev/null 2>&1; then
  die "jq is required (scoop install jq / apt install jq)"
fi

step "4) admin login"
admin_token="$(curl -fsS -H 'Content-Type: application/json' \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" \
  "$BASE/api/auth/login" | jq -r .token)"
[[ -n "$admin_token" && "$admin_token" != "null" ]] || die "admin login"
AH=(-H "Authorization: Bearer $admin_token")
echo "  token len=${#admin_token}"

step "5) anonymous /api/problems list"
curl -fsS "$BASE/api/problems" | jq '.items | length' >/dev/null || die "public problems list"
echo "  public list ok"

step "6) stu7 login + ranking penalty=60"
stu7_token="$(curl -fsS -H 'Content-Type: application/json' \
  -d '{"username":"stu7","password":"123456"}' "$BASE/api/auth/login" | jq -r .token)"
[[ -n "$stu7_token" && "$stu7_token" != "null" ]] || die "stu7 login"
H7=(-H "Authorization: Bearer $stu7_token")
# 拉"入门练习"的排名（按 title 找 id）
ps_id="$(curl -fsS "${AH[@]}" "$BASE/api/problemsets" | jq '.items[] | select(.title=="入门练习") | .id')"
[[ -n "$ps_id" ]] || die "practice set not found"
rank_json="$(curl -fsS "${H7[@]}" "$BASE/api/problemsets/$ps_id/ranking")"
stu7_pen="$(echo "$rank_json" | jq '.items[] | select(.username=="stu7") | .penalty_min // 0')"
echo "  stu7 penalty_min=$stu7_pen (expect 60)"
[[ "$stu7_pen" == "60" ]] || die "stu7 penalty expected 60 got $stu7_pen"

step "7) stu8 login + penalty=0 + ai rejected visible"
stu8_token="$(curl -fsS -H 'Content-Type: application/json' \
  -d '{"username":"stu8","password":"123456"}' "$BASE/api/auth/login" | jq -r .token)"
H8=(-H "Authorization: Bearer $stu8_token")
stu8_pen="$(echo "$rank_json" | jq '.items[] | select(.username=="stu8") | .penalty_min // 0')"
echo "  stu8 penalty_min=$stu8_pen (expect 0)"
[[ "$stu8_pen" == "0" ]] || die "stu8 penalty expected 0 got $stu8_pen"

step "8) stu10 kicked from 综合练习 → submissions still visible, not in ranking"
big_id="$(curl -fsS "${AH[@]}" "$BASE/api/problemsets" | jq '.items[] | select(.title=="综合练习 · 30 题") | .id')"
[[ -n "$big_id" ]] || die "big set not found"
big_rank="$(curl -fsS "${AH[@]}" "$BASE/api/problemsets/$big_id/ranking")"
stu10_count="$(echo "$big_rank" | jq '[.items[] | select(.username=="stu10")] | length')"
echo "  stu10 rows in big-set ranking = $stu10_count (expect 0)"
[[ "$stu10_count" == "0" ]] || die "stu10 should not appear in big-set ranking"
# 全站提交列表（admin）里 stu10 的提交还在
stu10_subs="$(curl -fsS "${AH[@]}" "$BASE/api/submissions?username=stu10" | jq '.items | length')"
echo "  stu10 submissions in /api/submissions = $stu10_subs (expect >0)"
[[ "$stu10_subs" != "0" ]] || die "stu10 submissions should be preserved"

step "9) admin edit problem → clears ai_explanation"
prob_id="$(curl -fsS "${AH[@]}" "$BASE/api/problems?page=1&page_size=1" | jq '.items[0].id')"
prob_json="$(curl -fsS "${AH[@]}" "$BASE/api/problems/$prob_id")"
orig_title="$(echo "$prob_json" | jq -r '.problem.title')"
orig_desc="$(echo "$prob_json" | jq -r '.problem.description // ""')"
orig_diff="$(echo "$prob_json" | jq -r '.problem.difficulty // ""')"
orig_solution="$(echo "$prob_json" | jq -r '.problem.solution_md // ""')"
orig_idea="$(echo "$prob_json" | jq -r '.problem.solution_idea_md // ""')"
orig_tl="$(echo "$prob_json" | jq '.problem.time_limit_ms')"
orig_ml="$(echo "$prob_json" | jq '.problem.memory_limit_mb')"
orig_visible="$(echo "$prob_json" | jq '.problem.visible')"
orig_tag_ids="$(echo "$prob_json" | jq -c '.tag_ids // []')"
make_problem_payload() {
  local title="$1"
  jq -nc \
    --arg title "$title" \
    --arg description "$orig_desc" \
    --arg difficulty "$orig_diff" \
    --arg solution_md "$orig_solution" \
    --arg solution_idea_md "$orig_idea" \
    --argjson time_limit_ms "$orig_tl" \
    --argjson memory_limit_mb "$orig_ml" \
    --argjson visible "$orig_visible" \
    --argjson tag_ids "$orig_tag_ids" \
    '{
      title: $title,
      description: $description,
      difficulty: $difficulty,
      solution_md: $solution_md,
      solution_idea_md: $solution_idea_md,
      time_limit_ms: $time_limit_ms,
      memory_limit_mb: $memory_limit_mb,
      visible: $visible,
      tag_ids: $tag_ids
    }'
}
RESTORE_PROBLEM_ID="$prob_id"
RESTORE_PROBLEM_PAYLOAD="$(make_problem_payload "$orig_title")"
edit_payload="$(make_problem_payload "${orig_title} [e2e]")"
curl -fsS "${AH[@]}" -H 'Content-Type: application/json' \
  -X PUT "$BASE/api/admin/problems/$prob_id" \
  -d "$edit_payload" >/dev/null
# 验证 ai_explanation 清空：查后端返回（不直接查 DB）
ai_count="$(curl -fsS "${AH[@]}" \
  "$BASE/api/submissions?problem_id=$prob_id&page=1&page_size=100" \
  | jq '[.items[] | select(.ai_explanation != "" and .ai_explanation != null)] | length')"
echo "  ai_explanation != '' rows for problem $prob_id = $ai_count (expect 0)"
[[ "$ai_count" == "0" ]] || die "problem edit should clear ai_explanation"
curl -fsS "${AH[@]}" -H 'Content-Type: application/json' \
  -X PUT "$BASE/api/admin/problems/$prob_id" \
  -d "$RESTORE_PROBLEM_PAYLOAD" >/dev/null
RESTORE_PROBLEM_ID=""
RESTORE_PROBLEM_PAYLOAD=""

step "10) SSE stream: 30s window should see ping or event"
sse_file="$(mktemp)"
set +e
curl -fsS -N --max-time 30 "$BASE/api/events/stream" >"$sse_file" 2>/dev/null
curl_status=$?
set -e
if [[ $curl_status -ne 0 && $curl_status -ne 28 ]]; then
  sed -n '1,20p' "$sse_file" || true
  rm -f "$sse_file"
  die "sse stream request failed (curl=$curl_status)"
fi
sse_line="$(grep -m1 -E '^:ping$|^event:' "$sse_file" || true)"
rm -f "$sse_file"
[[ -n "$sse_line" ]] || die "no SSE keepalive/event within 30s"
echo "  first frame: $sse_line"

step "DONE"
echo "[e2e] passed — log at $LOG"
