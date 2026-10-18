#!/bin/bash
# shell 资产自检：语法检查 + slog 功能测试

set -e

cd "$(dirname "$0")/.."

for f in lib/*.sh source/*.sh bin/*.sh tests/*.sh; do
    [ -f "$f" ] || continue
    bash -n "$f"
done

bash tests/sloggi.sh >/dev/null
bash tests/hallnet.sh >/dev/null

echo "[INFO] shell assets ok"
