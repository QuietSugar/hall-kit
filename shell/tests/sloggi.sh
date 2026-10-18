#!/bin/bash

# slog.sh 简单功能测试：验证各级日志能正确输出到 stdout

# shellcheck disable=SC1091
# shellcheck source=../lib/slog.sh
. "$(dirname "$0")/../lib/slog.sh"

LOG_LEVEL_STDOUT="DEBUG"

log "ok 1 - This is regular log message..."
log_info "ok 2 - So is this..."
log_success "ok 3 - Yeah!! Awesome Possum."
log_warning "ok 4 - Luke ... you turned off your targeting computer"
log_error "ok 5 - Whoops! I made a booboo"

# 测试 DEBUG 级别
log_debug "ok 6 - This should show up because LOG_LEVEL_STDOUT is DEBUG"

# 测试无效日志级别回退到 INFO
LOG_LEVEL_STDOUT="DOOTDOOT"
log_info "ok 7 - This should show up, even though LOG_LEVEL_STDOUT set to $LOG_LEVEL_STDOUT"

echo "[INFO] slog tests completed"
