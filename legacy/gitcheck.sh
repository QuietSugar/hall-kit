#!/bin/bash
# ====================================================
#   检查目标目录下所有 Git 仓库的本地状态
#   用于判断本地项目是否可以安全删除
#
#   用法: gitcheck [-a] [-d] [-q] [-r] [-t <目标目录>]
#     -a  同时打印 CLEAN 的项目
#     -d  调试模式
#     -q  安静模式，不打印配置和汇总信息
#     -r  开启远程状态检测：未推送 / 无上游分支 / 无远程
#     -t  指定目标目录，默认当前目录
#
#   退出码：发现 dirty 仓库时返回 1，否则返回 0
# ====================================================

# shellcheck disable=SC1091
# shellcheck source=lib/init.sh
. "$(dirname "$0")/lib/init.sh"

usage() {
    cat <<EOF
用法: $(basename "$0") [-a] [-d] [-q] [-r] [-t <目标目录>]
  -a  同时打印 CLEAN 的项目
  -d  调试模式
  -q  安静模式：不打印配置和汇总信息
  -r  开启远程状态检测：未推送 / 无上游分支 / 无远程
  -t  指定目标目录（默认当前目录）
EOF
}

print_clean=false
is_debug=false
quiet=false
check_remote=false
target_dir=$(pwd)
header_printed=false

while getopts "adt:qrh" opt; do
    case $opt in
        a) print_clean=true ;;
        d) is_debug=true ;;
        q) quiet=true ;;
        r) check_remote=true ;;
        t) target_dir=$(realpath_compat "$OPTARG") ;;
        h) usage; exit 0 ;;
        \?) usage >&2; exit 1 ;;
    esac
done
shift $((OPTIND - 1))

# 也支持直接传入目标目录作为位置参数
if [ $# -gt 0 ]; then
    case "$1" in
        http://*|https://*|git@*|ssh://*|ftp://*|file://*)
            log_error "gitcheck 用于检查本地 Git 仓库，不支持远程 URL: $1"
            log_info "用法: $(basename "$0") [-a] [-d] [-q] [-r] [-t <目标目录>]"
            exit 1
            ;;
    esac
    if ! target_dir=$(realpath_compat "$1"); then
        log_error "无效目录或路径不存在: $1"
        exit 1
    fi
fi

if [ ! -d "$target_dir" ]; then
    log_error "目录不存在: $target_dir"
    exit 1
fi

if [ "$quiet" != true ]; then
    log_info "此次检查目录: $target_dir"
fi

if [ "$is_debug" = true ]; then
    log_info "PRINT_CLEAN: $print_clean, CHECK_REMOTE: $check_remote"
fi

dirty_count=0
clean_count=0
exit_status=0

print_header() {
    if [ "$header_printed" = false ]; then
        printf '%b%-8s%b  %-12s  %s\n' \
            "$LOG_INFO_COLOR" "STATUS" "$LOG_DEFAULT_COLOR" "BRANCH" "PATH"
        header_printed=true
    fi
}

check_repo() {
    local repo_dir="$1"
    local relative_path="${repo_dir#"$target_dir"}"
    relative_path="${relative_path#/}"
    [ -z "$relative_path" ] && relative_path="."

    local status=""
    local branch
    branch=$(git -C "$repo_dir" rev-parse --abbrev-ref HEAD 2>/dev/null) || branch="?"

    # 未提交文件（包含 staged / unstaged / untracked）
    local unstaged
    unstaged=$(git -C "$repo_dir" status --short 2>/dev/null | wc -l | tr -d ' ')
    if [ "$unstaged" -ne 0 ]; then
        status+="[未提交${unstaged}]"
    fi

    # stash 数量
    local stash_count
    stash_count=$(git -C "$repo_dir" stash list 2>/dev/null | wc -l | tr -d ' ')
    if [ "$stash_count" -ne 0 ]; then
        status+="[储藏${stash_count}]"
    fi

    # 远程状态检测（默认不开启，避免与远程交互或依赖过时信息）
    if [ "$check_remote" = true ]; then
        local upstream
        if upstream=$(git -C "$repo_dir" rev-parse --abbrev-ref 'HEAD@{upstream}' 2>/dev/null); then
            local ahead
            ahead=$(git -C "$repo_dir" rev-list --count HEAD..."$upstream" --left-only 2>/dev/null | tr -d ' ')
            [ "$ahead" -ne 0 ] && status+="[未推送${ahead}]"
        elif [ -n "$(git -C "$repo_dir" remote -v 2>/dev/null)" ]; then
            status+="[无上游分支]"
        else
            status+="[无远程]"
        fi
    fi

    if [ -n "$status" ]; then
        print_header
        printf '%b%-8s%b  %-12s  %s\n' \
            "$LOG_WARN_COLOR" "[DIRTY]" "$LOG_DEFAULT_COLOR" "[$branch]" "$relative_path"
        printf '         %s\n' "$status"
        exit_status=1
        dirty_count=$((dirty_count + 1))
    else
        clean_count=$((clean_count + 1))
        if [ "$print_clean" = true ]; then
            print_header
            printf '%b%-8s%b  %-12s  %s\n' \
                "$LOG_SUCCESS_COLOR" "[CLEAN]" "$LOG_DEFAULT_COLOR" "[$branch]" "$relative_path"
        fi
    fi
}

# 查找所有最外层 Git 仓库（发现仓库根目录后不再继续深入其内部）
repo_count=0
while IFS= read -r -d '' repo_dir; do
    repo_count=$((repo_count + 1))
    if [ "$is_debug" = true ]; then
        log_info "[ DEBUG 仓库 ] $repo_dir"
    fi
    check_repo "$repo_dir"
done < <(find "$target_dir" -type d -exec test -d '{}/.git' \; -print0 -prune 2>/dev/null)

if [ "$repo_count" -eq 0 ]; then
    log_info "未找到 Git 仓库"
    exit 0
fi

if [ "$quiet" != true ]; then
    log_info "检查完成：dirty=${dirty_count}, clean=${clean_count}"
fi

exit "$exit_status"
