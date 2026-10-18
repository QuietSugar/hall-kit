# shell/ — 可复用的 bash 资产与 shell 工具

这里放 hall-kit 附带的 shell 内容；`hall shell install` 会把它们从二进制内解出：

- `lib/`、`source/`、`example.env` → `~/.hall/shell/`、`~/.hall/env.example`
- `bin/*.sh` → `~/.hall/bin/<name>`（去掉 `.sh`，加执行位）

## 命名约定：内置 vs 你写的

- `source/` 下**以 `hall-` 开头**的是内置片段：`install` 会覆盖更新、`uninstall` 会删除。
- 你自己写的片段也放 `~/.hall/shell/source/`，**不要用 `hall-` 前缀**。`install`/`uninstall` 只增删自己安装的文件（记录在 `~/.hall/shell/.state.json`），绝不碰你的文件。

## 编写约定

- 脚本统一 `#!/bin/bash`；纯 POSIX 的库可以保留 `#!/bin/sh`。
- 变量引用加双引号，路径拼接用 `"$base/$name"`，避免空格路径出错。
- `set -e` 谨慎使用，确保不会误伤正常流程。
- 目标环境是 Linux、macOS、Windows git-bash：避免平台特有命令，注意路径、权限、行尾符差异；仓库根的 `.gitattributes` 保证 `*.sh` 恒为 LF。

## lib/

- `slog.sh`：来自 [slog](http://github.com/swelljoe/slog)（MIT）的日志库，提供 `log_debug`/`log_info`/`log_success`/`log_warning`/`log_error`，级别由 `LOG_LEVEL_STDOUT`（默认 INFO）控制。
- `init.sh`：加载 `slog.sh` 与 `load_config.sh`，并提供 `realpath_compat`、`trim`。
- `hall-net.sh`：网络访问约定（`gh_url` / `gh_fetch`），直连不可用时统一走代理或 GitHub 加速前缀。
- `load_config.sh`：读取配置文件（`HALL_CONFIG_FILE` > `~/.hall/env` > 旧的 `~/.hall-command/env`），把每个 `key=value` 展开并 `export`。

库文件不会被自动加载，由脚本显式 source：

    . "$HOME/.hall/shell/lib/init.sh"
    log_info "hello"

## source/（登录时自动 source）

- `hall-env.sh`：把 `~/.hall/env` 读进当前 shell（内部 source `lib/load_config.sh`，提供静默的 `log_debug` 兜底）。
- `hall-proxy.sh`：`proxy [http://host:port]` 设置 http/https/all 代理（不带参数时用 `~/.hall/env` 里的 `PROXY_ADDR`）；`unproxy` 取消；保留 `set_proxy`/`unset_proxy` 兼容名。
- `hall-docker.sh`：`alias docker-compose='docker compose'`。

## bin/ — shell 工具

可执行脚本，安装后放进 `~/.hall/bin` 并由同一 marker 块加入 `PATH`（用 `case` 去重，重复 source 不会叠加）。

- `tat.sh` → `tat`：tmux 会话助手。按当前目录名 attach 或创建 tmux 会话；已在 tmux 中则切换过去。**需要本机安装 tmux**。

## 网络访问约定

直连 GitHub 不可用时，shell 侧统一用 `lib/hall-net.sh` 的 `gh_url` / `gh_fetch`：

- `https_proxy` / `http_proxy`：标准代理，curl 与 go 都认，优先级最高；
- `GH_PROXY`：GitHub 加速前缀（末尾斜杠可省），例如 `https://gh-proxy.com/`；默认空 = 直连，不默认走第三方镜像；
- `GOPROXY`：Go module 专用，例如 `https://goproxy.cn,direct`。

`gh_url` 只重写 GitHub 域（`raw.githubusercontent.com`、`github.com`、`codeload.github.com`），其它 URL 原样返回。

## 安装与管理

    go install github.com/QuietSugar/hall-kit/cmd/hall@latest
    hall shell install      # 解出资产、写 rc 块、必要时迁移旧配置
    hall shell status
    hall shell uninstall    # 只删自己装的文件，用户文件保留（--purge 才清空 ~/.hall）

## 测试

    bash shell/tests/run.sh

会先做全量 `bash -n`（lib/、source/、bin/、tests/），再运行 `tests/sloggi.sh` 验证各级日志输出。
