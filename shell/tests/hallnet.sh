#!/bin/bash
# hall-net.sh 简单测试：验证 gh_url 的重写规则

set -e

. "$(dirname "$0")/../lib/hall-net.sh"

fail=0
check() {
    if [ "$2" != "$3" ]; then
        echo "FAIL: $1 -> '$2' (want '$3')" >&2
        fail=1
    fi
}

check "direct raw" "$(gh_url "https://raw.githubusercontent.com/a/b")" "https://raw.githubusercontent.com/a/b"
check "direct repo" "$(gh_url "https://github.com/a/b")" "https://github.com/a/b"
check "other host untouched" "$(gh_url "https://example.com/x")" "https://example.com/x"

GH_PROXY="https://gh-proxy.com/"
check "mirror raw" "$(gh_url "https://raw.githubusercontent.com/a/b")" "https://gh-proxy.com/https://raw.githubusercontent.com/a/b"
check "mirror codeload" "$(gh_url "https://codeload.github.com/a/b/zip/x")" "https://gh-proxy.com/https://codeload.github.com/a/b/zip/x"

GH_PROXY="https://gh-proxy.com"
check "normalize missing slash" "$(gh_url "https://github.com/a/b")" "https://gh-proxy.com/https://github.com/a/b"

if [ "$fail" -eq 0 ]; then
    echo "[INFO] hall-net tests passed"
fi
exit "$fail"
