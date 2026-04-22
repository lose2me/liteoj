#!/usr/bin/env bash
# scripts/smoke-judge.sh —— 最小 go-judge 连通性探针。
# 只依赖 go-judge 本身，跟 liteoj 解耦；当学生反馈"判题挂了"时先用这个定位
# 到底是 OJ 问题还是沙箱问题。
#
# 前置：C:\WSL\start-go-judge.bat 已跑过，go-judge 在 127.0.0.1:5050 监听。

set -euo pipefail

JUDGE_URL="${JUDGE_URL:-http://127.0.0.1:5050}"

echo "[smoke-judge] GET $JUDGE_URL/version"
if ! body="$(curl -fsS "$JUDGE_URL/version" 2>&1)"; then
  echo "[smoke-judge] FAIL: $body"
  echo "[smoke-judge] HINT: run C:\\WSL\\start-go-judge.bat first"
  exit 1
fi
echo "[smoke-judge] version ok"
echo "$body" | head -c 200
echo

# 最小 run 请求：C++ Hello World
payload='{
  "cmd":[{
    "args":["/bin/sh","-c","echo Hello World"],
    "env":["PATH=/usr/bin:/bin"],
    "files":[
      {"content":""},
      {"name":"stdout","max":10240},
      {"name":"stderr","max":10240}
    ],
    "cpuLimit":1000000000,
    "memoryLimit":268435456,
    "procLimit":50
  }]
}'

echo "[smoke-judge] POST $JUDGE_URL/run (sh -c echo Hello World)"
resp="$(curl -fsS -H "Content-Type: application/json" -d "$payload" "$JUDGE_URL/run")"
echo "$resp" | head -c 400
echo
if echo "$resp" | grep -q '"status":"Accepted"'; then
  echo "[smoke-judge] OK — sandbox returned Accepted"
  exit 0
fi
echo "[smoke-judge] FAIL — sandbox did not return Accepted"
exit 1
