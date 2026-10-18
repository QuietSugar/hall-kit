# Changelog

本项目遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

每次发布时：把 `[Unreleased]` 整理为对应版本号章节（含日期），并在其顶部保留新的空 `[Unreleased]` 章节。

## [Unreleased]

## [0.0.1] - 2026-10-18

### Added

- `hall` 工具箱总入口 + `hall aliases` 短名（同二进制软链 / git-bash shim）：`dockerage`、`gitcheck`、`gitclone`、`gitarc`、`rmempty`；`cmd/<tool>` 保留兼容 shim
- `hall shell` 系列：`install`（解出资产到 ~/.hall/shell/，幂等写 rc 块）/`status`/`uninstall`/`migrate`（导回旧 ~/.hall-command 配置，`--purge` 带时间戳备份）；配置 `~/.hall/env` 模板生成；内置 `hall-` 片段（env/proxy/docker）登录自动 source
- `tat` tmux 会话助手；`proxy` / `unproxy` 代理开关片段
- Windows（git-bash + Windows Terminal）安装 shim 支持（实验性）
- `install.sh` 引导：release tag 解析（`HALL_KIT_VERSION`，默认 latest）+ `GH_PROXY` 加速 + `HALL_KIT_INSTALL_DIR` 安装 `~/.hall/bin`
- 工程化：CI（shellcheck + golangci-lint）、tag 触发 release 产出多平台二进制
