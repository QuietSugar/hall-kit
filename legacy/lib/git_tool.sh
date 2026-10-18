#!/bin/bash
# ====================================================

# 判断字符串是否为 HTTPS/HTTP 或 SSH 协议的有效 Git 地址（支持端口和 IP）
is_git_url_https_ssh() {
    local url="$1"
    # 正则表达式：
    # 1. HTTPS/HTTP：支持域名/IP + 可选端口 + 路径 + 可选.git后缀
    # 2. SSH：支持用户名@域名/IP + 可选端口 + 路径 + 可选.git后缀
    local git_regex='^(https?)://([a-zA-Z0-9.-]+|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]{1,5})?(/[a-zA-Z0-9._/-]+)*(\.git)?$|^[a-zA-Z0-9._-]+@([a-zA-Z0-9.-]+|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]{1,5})?:[a-zA-Z0-9._/-]+(\.git)?$'

    if [[ "$url" =~ $git_regex ]]; then
        return 0 # 有效
    else
        log_error "无效的 Git 地址: $url"
        return 1
    fi
}

get_git_repo_path() {
  # 目标基础目录，支持通过环境变量覆盖
  local git_repo_path="${HALL_GIT_REPO_PATH:-${HOME}/git-repo}"
  if [ ! -d "$git_repo_path" ]; then
      mkdir -p "$git_repo_path"
  fi
  echo "$git_repo_path"
}

# 将 git url 转化成目录 规范化文件名 尽量使得三端都支持
#  http://1.2.3.4:8080/a/b/c.git -> 1.2.3.4/8080/a/b/c
#  git@1.2.3.4:a/b/c.git -> 1.2.3.4/a/b/c
make_filename_safe() {
    echo "$1" | sed \
        -e 's/^[[:space:]]*//' \
        -e 's/[[:space:]]*$//' \
        -e 's|^https://gh-proxy.com/||' \
        -e 's|^https://||' \
        -e 's|^http://||' \
        -e 's|^git@||' \
        -e 's/:/\//g' \
        -e 's/\.git$//'
}
