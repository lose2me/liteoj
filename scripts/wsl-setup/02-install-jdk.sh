#!/usr/bin/env bash
# 02-install-jdk.sh — 安装 Temurin OpenJDK 17 (tarball 方式, ghfast.top 代理)
set -euo pipefail

JDK_MAJOR=17
JDK_VERSION="17.0.13+11"
JDK_URL_VERSION="jdk-17.0.13%2B11"
JDK_TARBALL="OpenJDK17U-jdk_x64_linux_hotspot_17.0.13_11.tar.gz"
JDK_ROOT="/opt/java"
JDK_DIR="${JDK_ROOT}/temurin-17"

# 下载代理链 (按顺序尝试)
URLS=(
  "https://ghfast.top/https://github.com/adoptium/temurin17-binaries/releases/download/${JDK_URL_VERSION}/${JDK_TARBALL}"
  "https://mirror.ghproxy.com/https://github.com/adoptium/temurin17-binaries/releases/download/${JDK_URL_VERSION}/${JDK_TARBALL}"
  "https://github.com/adoptium/temurin17-binaries/releases/download/${JDK_URL_VERSION}/${JDK_TARBALL}"
)

mkdir -p "${JDK_ROOT}"

if [ ! -x "${JDK_DIR}/bin/java" ]; then
    cd /tmp
    rm -f "${JDK_TARBALL}"
    for URL in "${URLS[@]}"; do
        echo "==> 尝试: ${URL}"
        if curl -fsSL --retry 2 --connect-timeout 10 -o "${JDK_TARBALL}" "${URL}"; then
            if [ -s "${JDK_TARBALL}" ]; then
                echo "==> 下载成功 ($(stat -c%s "${JDK_TARBALL}") bytes)"
                break
            fi
        fi
        rm -f "${JDK_TARBALL}"
    done

    if [ ! -s "${JDK_TARBALL}" ]; then
        echo "!! 所有源均失败"; exit 1
    fi

    echo "==> 解压到 ${JDK_ROOT}"
    tar -xzf "${JDK_TARBALL}" -C "${JDK_ROOT}/"
    # 解压出 jdk-17.0.13+11 目录, rename 为 temurin-17
    rm -rf "${JDK_DIR}"
    mv "${JDK_ROOT}"/jdk-17.* "${JDK_DIR}"
    rm -f "${JDK_TARBALL}"
fi

echo "==> 注册 update-alternatives"
update-alternatives --install /usr/bin/java  java  "${JDK_DIR}/bin/java"  100
update-alternatives --install /usr/bin/javac javac "${JDK_DIR}/bin/javac" 100
update-alternatives --install /usr/bin/jar   jar   "${JDK_DIR}/bin/jar"   100
update-alternatives --set java  "${JDK_DIR}/bin/java"
update-alternatives --set javac "${JDK_DIR}/bin/javac"
update-alternatives --set jar   "${JDK_DIR}/bin/jar"

echo "==> 写 /etc/profile.d/java.sh"
cat > /etc/profile.d/java.sh <<EOF
export JAVA_HOME=${JDK_DIR}
export PATH=\$JAVA_HOME/bin:\$PATH
EOF
chmod 644 /etc/profile.d/java.sh

echo "==> Java 验证:"
java -version
javac -version
readlink -f "$(command -v java)"
