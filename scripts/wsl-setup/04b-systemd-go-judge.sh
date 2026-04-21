#!/usr/bin/env bash
# 04b-systemd-go-judge.sh — 手写 systemd unit 并启动 go-judge, 监听 127.0.0.1:5050
set -euo pipefail

echo "==> 清理旧的 drop-in (若存在)"
rm -rf /etc/systemd/system/go-judge.service.d

echo "==> 写 /etc/systemd/system/go-judge.service"
cat > /etc/systemd/system/go-judge.service <<'EOF'
[Unit]
Description=go-judge sandbox service for LiteOJ
Documentation=https://github.com/criyle/go-judge
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/bin/go-judge -http-addr=127.0.0.1:5050 -enable-grpc=false -silent=false
Restart=on-failure
RestartSec=3
LimitNOFILE=65536
# go-judge 需要 cgroup / ns 相关能力
AmbientCapabilities=CAP_SYS_ADMIN CAP_SYS_PTRACE CAP_SYS_RESOURCE CAP_DAC_OVERRIDE CAP_CHOWN CAP_FOWNER CAP_SETGID CAP_SETUID
NoNewPrivileges=false

[Install]
WantedBy=multi-user.target
EOF

echo "==> daemon-reload & enable & restart"
systemctl daemon-reload
systemctl enable go-judge
systemctl restart go-judge

# 等待就绪
for i in 1 2 3 4 5 6 7 8 9 10; do
    if ss -tln 2>/dev/null | grep -q ':5050'; then
        echo "==> 端口 5050 已监听 (第 ${i} 秒)"
        break
    fi
    sleep 1
done

echo "==> systemctl status:"
systemctl --no-pager status go-judge | head -25 || true

echo "==> 端口监听详情:"
ss -tlnp 2>/dev/null | grep 5050 || ss -tln | grep 5050 || echo "!!! 端口 5050 未监听"

echo "==> 本地 curl 测试:"
curl -fsS http://127.0.0.1:5050/version && echo
curl -fsS http://127.0.0.1:5050/config | head -c 300 && echo
