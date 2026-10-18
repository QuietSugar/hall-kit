#!/bin/bash
# hall-kit built-in snippet: proxy helpers.
#
# proxy [http://host:port]   set http/https/all proxy; without an argument the
#                            PROXY_ADDR value from ~/.hall/env is used
# unproxy                    clear the proxy
# set_proxy / unset_proxy    compatibility aliases for the old host config

proxy() {
    local addr="${1:-$PROXY_ADDR}"
    if [ -z "$addr" ]; then
        echo "用法: proxy <http://host:port>（或在 ~/.hall/env 设置 PROXY_ADDR）" >&2
        return 1
    fi
    case "$addr" in
        http://*|https://*) ;;
        *) addr="http://$addr" ;;
    esac
    export http_proxy="$addr" https_proxy="$addr" all_proxy="$addr"
    echo "代理已设置为: $addr"
}

unproxy() {
    unset http_proxy https_proxy all_proxy
    echo "代理已取消"
}

set_proxy()   { proxy; }
unset_proxy() { unproxy; }
