#!/bin/bash
#--------------------------------------------------------------------------------------------------
# slog - Makes logging in shell scripting suck less
# Copyright (c) Fred Palmer
# Licensed under the MIT license
# http://github.com/swelljoe/slog
#--------------------------------------------------------------------------------------------------

# LOG_LEVEL_STDOUT - Define to determine above which level goes to STDOUT.
# By default, all log levels will be written to STDOUT.
LOG_LEVEL_STDOUT="INFO"

# Useful global variables that users may wish to reference
SCRIPT_NAME="$0"
SCRIPT_NAME="${SCRIPT_NAME#\./}"
SCRIPT_NAME="${SCRIPT_NAME##/*/}"

# Determines if we print colors or not
if tty -s >/dev/null 2>&1; then
    readonly INTERACTIVE_MODE="on"
else
    readonly INTERACTIVE_MODE="off"
fi

#--------------------------------------------------------------------------------------------------
# Begin Logging Section
if [ "${INTERACTIVE_MODE}" = "off" ]
then
    # Then we don't care about log colors
    LOG_DEFAULT_COLOR=""
    LOG_ERROR_COLOR=""
    LOG_INFO_COLOR=""
    LOG_SUCCESS_COLOR=""
    LOG_WARN_COLOR=""
    LOG_DEBUG_COLOR=""
else
    LOG_DEFAULT_COLOR=$(tput sgr0)
    LOG_ERROR_COLOR=$(tput setaf 1)
    LOG_INFO_COLOR=$(tput sgr 0)
    LOG_SUCCESS_COLOR=$(tput setaf 2)
    LOG_WARN_COLOR=$(tput setaf 3)
    LOG_DEBUG_COLOR=$(tput setaf 4)
fi

printf '%s[%s]%s\n' "$LOG_SUCCESS_COLOR" "$(date +"%Y-%m-%d %H:%M:%S %Z")" "$LOG_DEFAULT_COLOR"

# 将日志级别名称转换为整数
log_level_to_int() {
    case "$1" in
        DEBUG)   echo 0 ;;
        INFO)    echo 1 ;;
        SUCCESS) echo 2 ;;
        WARNING) echo 3 ;;
        ERROR)   echo 4 ;;
        *)       echo 1 ;;
    esac
}

log() {
    local log_text="$1"
    local log_level="$2"
    local log_color="$3"

    # Default level to "info"
    [ -z "${log_level}" ] && log_level="INFO"
    [ -z "${log_color}" ] && log_color="${LOG_INFO_COLOR}"

    # Validate LOG_LEVEL_STDOUT
    case $LOG_LEVEL_STDOUT in
        DEBUG|INFO|SUCCESS|WARNING|ERROR)
            ;;
        *)
            LOG_LEVEL_STDOUT=INFO
            ;;
    esac

    # Check LOG_LEVEL_STDOUT to see if this level of entry goes to STDOUT.
    local log_level_int log_level_stdout
    log_level_int=$(log_level_to_int "$log_level")
    log_level_stdout=$(log_level_to_int "$LOG_LEVEL_STDOUT")
    if [ "$log_level_stdout" -le "$log_level_int" ]; then
        # STDOUT
        printf '%s[%s] %s %s\n' "$log_color" "$(date +"%H:%M:%S")" "$log_text" "$LOG_DEFAULT_COLOR"
    fi
    return 0
}

log_info()      { log "$@"; }
log_success()   { log "$1" "SUCCESS" "${LOG_SUCCESS_COLOR}"; }
log_error()     { log "$1" "ERROR" "${LOG_ERROR_COLOR}"; }
log_warning()   { log "$1" "WARNING" "${LOG_WARN_COLOR}"; }
log_debug()     { log "$1" "DEBUG" "${LOG_DEBUG_COLOR}"; }

# End Logging Section
#--------------------------------------------------------------------------------------------------
