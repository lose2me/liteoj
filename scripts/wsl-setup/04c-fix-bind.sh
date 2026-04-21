#!/usr/bin/env bash
# 04c-fix-bind.sh — 改 go-judge 绑 0.0.0.0 (让 WSL localhost forwarding 生效), 从外部验证
set -euo pipefail

echo "==> 改 systemd unit 绑 0.0.0.0:5050 (Windows 端仍只暴露 127.0.0.1:5050, 因为 WSL2 NAT 只转发到 Windows loopback)"
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
# -http-addr=0.0.0.0:5050 让 WSL2 localhost-forwarding 把它映射到 Windows 127.0.0.1:5050
ExecStart=/usr/bin/go-judge -http-addr=0.0.0.0:5050 -enable-grpc=false -silent=false
Restart=on-failure
RestartSec=3
LimitNOFILE=65536
AmbientCapabilities=CAP_SYS_ADMIN CAP_SYS_PTRACE CAP_SYS_RESOURCE CAP_DAC_OVERRIDE CAP_CHOWN CAP_FOWNER CAP_SETGID CAP_SETUID
NoNewPrivileges=false

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl restart go-judge
sleep 2

echo "==> 监听详情:"
ss -tlnp | grep 5050

echo "==> WSL 内自测:"
curl -fsS http://127.0.0.1:5050/version && echo
