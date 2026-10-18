#!/bin/bash
# hall-kit built-in snippet: load ~/.hall/env into the current shell.
#
# Config precedence is handled by load_config.sh:
#   HALL_CONFIG_FILE > ~/.hall/env > ~/.hall-command/env (legacy)
#
# load_config.sh logs through log_debug from slog.sh; provide a silent stub so
# this snippet does not drag in slog.sh (which prints a banner line when
# sourced).

if ! command -v log_debug >/dev/null 2>&1; then
    log_debug() { :; }
fi

if [ -f "$HOME/.hall/shell/lib/load_config.sh" ]; then
    . "$HOME/.hall/shell/lib/load_config.sh"
fi
