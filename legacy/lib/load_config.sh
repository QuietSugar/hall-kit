#!/bin/bash


## @description:
# ====================================================
#   初始化变量,用于其他脚本调用,配置文件统一放在 $HOME/.hall-command/ 下
#   配置文件内容的的格式是 key=value
#   如何使用: 由 init.sh 自动加载
#
# ====================================================

config_file="${HOME}/.hall-command/env"

if [ -f "$config_file" ]; then
    log_debug "--->>> load env from: $config_file"
    while IFS= read -r line || [ -n "$line" ]; do
        # 去除首尾空白
        line=$(printf '%s' "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        # 跳过空行与注释
        [ -z "$line" ] && continue
        case "$line" in
            \#*) continue ;;
        esac
        key=${line%%=*}
        value=${line#*=}
        [ -z "$key" ] && continue
        # 扩展配置文件中常见的路径变量
        value="${value/#\~/$HOME}"
        value="${value//\$\{HOME\}/$HOME}"
        value="${value//\$HOME/$HOME}"
        export "$key=$value"
        log_debug "$key=$value"
    done < "$config_file"
    log_debug "--->>> load env from : finish load"
else
    log_debug "文件不存在 $config_file"
fi


