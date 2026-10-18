#!/bin/bash

# ====================================================
# 初始化
# 1. 加载日志组件
# 2. 加载配置文件
# 3. 提供跨平台工具函数
#
# ====================================================

# 确定脚本的规范路径
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
    DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
    SOURCE="$(readlink "$SOURCE")"
    [[ "$SOURCE" != /* ]] && SOURCE="$DIR/$SOURCE"
done
# 脚本所在目录的路径
SCRIPT_DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# shellcheck disable=SC1091
# shellcheck source=command/lib/slog.sh
. "$SCRIPT_DIR/slog.sh"
# shellcheck disable=SC1091
# shellcheck source=command/lib/load_config.sh
. "$SCRIPT_DIR/load_config.sh"

# 跨平台 realpath 兼容实现
# macOS 默认没有 realpath，用 cd + pwd 实现等效功能
realpath_compat() {
    local path="$1"
    if [ -d "$path" ]; then
        (cd -P -- "$path" && pwd -P)
    elif [ -e "$path" ]; then
        local dir
        dir=$(cd -P -- "$(dirname -- "$path")" && pwd -P)
        printf '%s/%s\n' "$dir" "$(basename -- "$path")"
    else
        return 1
    fi
}

trim() {
    local var="$1"
    var="${var#"${var%%[![:space:]]*}"}"  # 去除开头空白
    var="${var%"${var##*[![:space:]]}"}"  # 去除结尾空白
    printf '%s' "$var"
}
