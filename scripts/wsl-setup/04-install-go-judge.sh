#!/usr/bin/env bash
# 04-install-go-judge.sh — 安装 go-judge (.deb) 并配置 systemd 监听 127.0.0.1:5050
set -euo pipefail

GO_JUDGE_VERSION="1.11.4"
DEB_NAME="go-judge_${GO_JUDGE_VERSION}_linux_amd64v2.deb"
URLS=(
  "https://ghfast.top/https://github.com/criyle/go-judge/releases/download/v${GO_JUDGE_VERSION}/${DEB_NAME}"
  "https://mirror.ghproxy.com/https://github.com/criyle/go-judge/releases/download/v${GO_JUDGE_VERSION}/${DEB_NAME}"
  "https://github.com/criyle/go-judge/releases/download/v${GO_JUDGE_VERSION}/${DEB_NAME}"
)

cd /tmp
rm -f "${DEB_NAME}"
echo "==> 下载 go-judge v${GO_JUDGE_VERSION} (.deb)"
for URL in "${URLS[@]}"; do
    echo "  尝试: ${URL}"
    if curl -fsSL --retry 2 --connect-timeout 10 -o "${DEB_NAME}" "${URL}" && [ -s "${DEB_NAME}" ]; then
        SIZE=$(stat -c%s "${DEB_NAME}")
        if [ "${SIZE}" -gt 1000000 ]; then
            echo "  成功, 大小 ${SIZE} bytes"
            break
        fi
        echo "  文件过小 (${SIZE} bytes), 可能是重定向页, 重试下一个源"
        rm -f "${DEB_NAME}"
    fi
done

if [ ! -s "${DEB_NAME}" ]; then
    echo "!! 所有源均失败"; exit 1
fi

echo "==> 安装 deb"
apt-get install -y "./${DEB_NAME}"
rm -f "${DEB_NAME}"

echo "==> 查看 deb 提供的文件"
dpkg -L go-judge | head -30
echo "---"

# 检查默认 systemd 单元
if [ -f /lib/systemd/system/go-judge.service ]; then
    echo "==> 发现默认 systemd unit: /lib/systemd/system/go-judge.service"
    cat /lib/systemd/system/go-judge.service
fi

echo "==> 写入自定义 go-judge 配置 /etc/go-judge/mount.yaml (如 deb 已装则复用)"
mkdir -p /etc/go-judge
# 走 drop-in 方式复写 ExecStart 指定监听和端口
mkdir -p /etc/systemd/system/go-judge.service.d
cat > /etc/systemd/system/go-judge.service.d/override.conf <<'EOF'
[Service]
# 清空原 ExecStart, 再指定我们的参数: 仅监听 loopback, 端口 5050
ExecStart=
ExecStart=/usr/bin/go-judge -http-addr=127.0.0.1:5050 -enable-grpc=false -silent=false
Restart=on-failure
RestartSec=3
EOF

echo "==> 重载 systemd 并启动"
systemctl daemon-reload
systemctl enable go-judge
systemctl restart go-judge

sleep 2
echo "==> go-judge 状态:"
systemctl status go-judge --no-pager -l | head -30 || true

echo "==> 端口监听:"
ss -tlnp 2>/dev/null | grep 5050 || ss -tln | grep 5050 || echo "端口 5050 还未监听, 等下再看"

echo "==> 本地 curl 测试:"
sleep 2
curl -fsS http://127.0.0.1:5050/version && echo || echo "GET /version failed"
curl -fsS http://127.0.0.1:5050/config | head -c 200 && echo || echo "GET /config failed"
