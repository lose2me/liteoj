#!/usr/bin/env bash
# 01-bootstrap.sh — 写入 /etc/wsl.conf 并切换 apt 到清华源。
set -euo pipefail

echo "==> 写入 /etc/wsl.conf (启用 systemd)"
cat > /etc/wsl.conf <<'EOF'
[boot]
systemd=true

[user]
default=root

[network]
hostname=liteoj
generateHosts=true
generateResolvConf=true

[interop]
enabled=true
appendWindowsPath=false
EOF

echo "==> 备份并替换 apt sources (trixie → 清华 tuna, 先 HTTP 再切 HTTPS)"
if [ -f /etc/apt/sources.list.d/debian.sources ]; then
    cp /etc/apt/sources.list.d/debian.sources /etc/apt/sources.list.d/debian.sources.bak
fi
# 第一阶段：HTTP 源，用来装 ca-certificates
cat > /etc/apt/sources.list.d/debian.sources <<'EOF'
Types: deb
URIs: http://mirrors.tuna.tsinghua.edu.cn/debian
Suites: trixie trixie-updates trixie-backports
Components: main contrib non-free non-free-firmware
Signed-By: /usr/share/keyrings/debian-archive-keyring.gpg

Types: deb
URIs: http://mirrors.tuna.tsinghua.edu.cn/debian-security
Suites: trixie-security
Components: main contrib non-free non-free-firmware
Signed-By: /usr/share/keyrings/debian-archive-keyring.gpg
EOF

# 注释掉老的 sources.list (若有)
if [ -f /etc/apt/sources.list ]; then
    sed -i 's|^deb |# deb |' /etc/apt/sources.list || true
fi

echo "==> apt update (HTTP 阶段)"
apt-get update

echo "==> 先装 ca-certificates (为切到 HTTPS 做准备)"
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates

echo "==> 切换到 HTTPS 源"
sed -i 's|http://mirrors.tuna.tsinghua.edu.cn|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources

echo "==> apt update (HTTPS 阶段)"
apt-get update

echo "==> apt upgrade"
DEBIAN_FRONTEND=noninteractive apt-get -y upgrade

echo "==> 安装基础工具"
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates curl wget git vim less procps \
    build-essential pkg-config \
    libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev \
    libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev \
    libffi-dev liblzma-dev uuid-dev

echo "==> bootstrap 完成"
