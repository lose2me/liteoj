#!/usr/bin/env bash
# 03-install-python311.sh — 通过 pyenv 安装 Python 3.11，并做系统链接。
set -euo pipefail

PYENV_ROOT=/opt/pyenv
PY_VERSION=3.11.11

export PYENV_ROOT
export PATH="${PYENV_ROOT}/bin:${PATH}"
# 清华 python 源镜像，加速 pyenv 下载
export PYTHON_BUILD_MIRROR_URL="https://mirrors.tuna.tsinghua.edu.cn/python"
export PYTHON_BUILD_MIRROR_URL_SKIP_CHECKSUM=0

echo "==> 克隆 pyenv 到 ${PYENV_ROOT} (走 ghfast.top 代理加速)"
if [ ! -d "${PYENV_ROOT}" ]; then
    git clone --depth=1 https://ghfast.top/https://github.com/pyenv/pyenv.git "${PYENV_ROOT}" \
        || git clone --depth=1 https://gitee.com/mirrors/pyenv.git "${PYENV_ROOT}"
fi

echo "==> pyenv 版本:"
"${PYENV_ROOT}/bin/pyenv" --version

echo "==> 写 /etc/profile.d/pyenv.sh"
cat > /etc/profile.d/pyenv.sh <<'EOF'
export PYENV_ROOT=/opt/pyenv
export PATH="$PYENV_ROOT/bin:$PATH"
if command -v pyenv >/dev/null 2>&1; then
    eval "$(pyenv init -)"
fi
EOF
chmod 644 /etc/profile.d/pyenv.sh

echo "==> 安装 Python ${PY_VERSION} (从 ${PYTHON_BUILD_MIRROR_URL} 下载源码构建)"
if ! "${PYENV_ROOT}/bin/pyenv" versions --bare | grep -qx "${PY_VERSION}"; then
    "${PYENV_ROOT}/bin/pyenv" install -s "${PY_VERSION}"
fi
"${PYENV_ROOT}/bin/pyenv" global "${PY_VERSION}"

PYTHON311_BIN="${PYENV_ROOT}/versions/${PY_VERSION}/bin/python3.11"

echo "==> 建立软链 /usr/local/bin/python3.11 -> ${PYTHON311_BIN}"
ln -sf "${PYTHON311_BIN}" /usr/local/bin/python3.11
ln -sf "${PYTHON311_BIN}" /usr/local/bin/python3

echo "==> 配置 pip 清华镜像"
"${PYTHON311_BIN}" -m pip config --global set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
"${PYTHON311_BIN}" -m pip config --global set install.trusted-host pypi.tuna.tsinghua.edu.cn

echo "==> Python 版本验证:"
python3 --version
python3.11 --version
which python3
which python3.11
