#!/bin/bash
# ====================================================
#   遍历寻找目录下所有带 .git的目录,
#	将它归档到 固定目录中
#
# ====================================================

# shellcheck disable=SC1091
# shellcheck source=lib/init.sh
. "$(dirname "$0")/lib/init.sh"
# shellcheck disable=SC1091
# shellcheck source=lib/git_tool.sh
. "$(dirname "$0")/lib/git_tool.sh"

# shellcheck disable=SC2034
LOG_LEVEL_STDOUT="INFO"

set -e
# 基准目录(也就是扫描的根目录)
if [ $# -eq 0 ]; then
    SOURCE_BASE_ABS_PATH=$(realpath_compat ".")
    log_success "此次操作当前目录: $SOURCE_BASE_ABS_PATH"
else
    SOURCE_BASE_ABS_PATH=$(realpath_compat "$1")
    log_success "此次操作指定目录: $SOURCE_BASE_ABS_PATH"
fi

# 检查目录是否存在
if [ ! -d "$SOURCE_BASE_ABS_PATH" ]; then
    log_error "Error: 目录 $SOURCE_BASE_ABS_PATH 不存在"
    exit 1
fi
# 目标基础目录
TARGET_BASE_ABS_PATH=$(get_git_repo_path)

# 查找所有.git目录的函数
find_git_dirs() {
    if [ $# -eq 0 ]; then
        log_error "请传入参数!"
        exit 1
    fi
    local this_dir="$1"
    # 检查当前目录是否有.git子目录
    if [ -d "$this_dir/.git" ]; then
        log_debug "【原始】绝对路径: $this_dir"
        move_git_dir "$this_dir"
        log_success ""
        return
    fi

    # 遍历当前目录的所有子目录
    for subdir in "$this_dir"/*; do
        # 如果是目录且不是符号链接（避免循环）
        if [ -d "$subdir" ] && [ ! -L "$subdir" ]; then
            find_git_dirs "$subdir"
        fi
    done
}
move_git_dir() {
    # eg /base/a/b/c
    local source_abs_dir="$1"
    # eg /base/a/b/c/.git
    local source_dot_git_abs_dir="$source_abs_dir/.git"
    # eg http://1.2.3.4/a/b/c.git
    local cloneUrl
    cloneUrl=$(git --git-dir="$source_dot_git_abs_dir" config --get remote.origin.url || true)
    if ! is_git_url_https_ssh "$cloneUrl"; then
        log_warning "跳过无效 Git 地址: $cloneUrl"
        return
    fi
    #替换
    local project_dir
    project_dir=$(make_filename_safe "$cloneUrl")
    case "/$project_dir/" in
        */../*)
            log_warning "跳过不安全的项目路径: $project_dir"
            return
            ;;
    esac
    log_debug "【目标】项目相对路径:  $project_dir"
    # eg c
    local project_name
    project_name="${project_dir##*/}"
    log_debug "【目标】项目名称:  $project_name"
    # 不带前缀和后缀的 git url eg 1.2.3.4/a/b/c
    local git_clone_to_dir
    git_clone_to_dir=${project_dir%/*}
    log_debug "【目标】所在目录相对路径:  $git_clone_to_dir"
    local absolute_project_dir
    absolute_project_dir="$TARGET_BASE_ABS_PATH/$git_clone_to_dir"
    log_debug "【目标】所在目录绝对路径:  $absolute_project_dir"

    if [ ! -d "$absolute_project_dir" ]; then
        mkdir -p "$absolute_project_dir"
    fi
    local new_path="$absolute_project_dir/$project_name"
    log_debug "【new_path:  $new_path"
    if [ ! -d "$new_path" ]; then
        # 使用mv实现移动和重命名
        log_success "从:              $source_abs_dir"
        log_success "移动并重命名为   $new_path"
        mv -v "$source_abs_dir" "$new_path"
    else
        log_warning "目录已存在,无法移动$new_path"
    fi
}



# 调用函数开始查找
find_git_dirs "$SOURCE_BASE_ABS_PATH"
