#!/bin/sh
# hall-kit 网络访问约定（shell 侧共用实现）
#
# 变量（两个项目同名同义）：
#   http_proxy / https_proxy  标准代理；curl 与 go 都认，优先级最高
#   GH_PROXY                  GitHub 加速前缀，例如 https://gh-proxy.com/
#                             （末尾斜杠可省；默认空 = 直连，不默认走第三方镜像）
#   GOPROXY                   Go module 专用，例如 https://goproxy.cn,direct
#
#   gh_url <url>            → 打印实际要访问的 URL（只重写 GitHub 域）
#   gh_fetch <url> [dest]   → 用 curl/wget 下载（自动套用 gh_url）

gh_url() {
    case "$1" in
        https://raw.githubusercontent.com/*|https://github.com/*|https://codeload.github.com/*)
            if [ -n "$GH_PROXY" ]; then
                printf '%s%s\n' "${GH_PROXY%/}/" "$1"
            else
                printf '%s\n' "$1"
            fi
            ;;
        *)
            printf '%s\n' "$1"
            ;;
    esac
}

gh_fetch() {
    _hall_url=$(gh_url "$1") || return 1
    if command -v curl >/dev/null 2>&1; then
        if [ -n "$2" ]; then
            curl -fL --connect-timeout 10 --max-time 60 --retry 2 -o "$2" "$_hall_url"
        else
            curl -fsSL --connect-timeout 10 --max-time 60 --retry 2 "$_hall_url"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if [ -n "$2" ]; then
            wget --timeout=60 --tries=3 -O "$2" "$_hall_url"
        else
            wget -nv --timeout=60 --tries=3 -O - "$_hall_url"
        fi
    else
        echo "gh_fetch: curl/wget not found" >&2
        return 1
    fi
}
