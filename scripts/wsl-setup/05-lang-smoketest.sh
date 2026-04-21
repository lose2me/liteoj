#!/usr/bin/env bash
# 05-lang-smoketest.sh — 在 WSL 里验证四种语言的工具链版本
set -uo pipefail  # 故意不 set -e, 因测试程序会返回 42/17

echo "=== C (gcc) ==="
gcc --version | head -1
echo 'int main(void){return 42;}' > /tmp/t.c
gcc -std=c11 -O2 -o /tmp/t.c.out /tmp/t.c; /tmp/t.c.out; echo "  exit=$? (expect 42)"

echo
echo "=== C++ (g++) ==="
g++ --version | head -1
echo 'int main(){return 17;}' > /tmp/t.cc
g++ -std=c++17 -O2 -o /tmp/t.cc.out /tmp/t.cc; /tmp/t.cc.out; echo "  exit=$? (expect 17)"

echo
echo "=== Python 3.11 ==="
/usr/local/bin/python3.11 --version
/usr/local/bin/python3.11 -c "import sys; print('py ok from', sys.executable)"

echo
echo "=== Java 17 ==="
java -version 2>&1 | head -2
javac -version 2>&1 | head -1
cat > /tmp/H.java <<'J'
public class H { public static void main(String[] a){ System.out.println("java ok"); } }
J
javac -d /tmp /tmp/H.java && (cd /tmp && java H)

echo
echo "=== Sandbox (go-judge) internal probe ==="
curl -fsS http://127.0.0.1:5050/version | head -c 200; echo
curl -fsS http://127.0.0.1:5050/config | /usr/local/bin/python3.11 -c "import sys,json; c=json.load(sys.stdin); rc=c.get('runnerConfig',{}); print('cgroupType:', rc.get('cgroupType')); print('cgroupControllers:', rc.get('cgroupControllers'))"
