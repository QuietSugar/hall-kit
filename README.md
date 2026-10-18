# hall-kit

hall 的日常 Go CLI 工具箱：发布**一个**多调用二进制 `hall`，各工具以子命令（`hall <tool>`）或短名（`gitcheck` 等）调用。

## 安装

预编译二进制（无需 Go）：

```bash
curl -fsSL https://raw.githubusercontent.com/QuietSugar/hall-kit/master/install.sh | sh
```

访问不了 GitHub 时，用加速前缀拉取脚本（脚本下载资产时会自动使用同一个加速前缀，无需再设变量）：

```bash
curl -fsSL https://gh-proxy.com/https://raw.githubusercontent.com/QuietSugar/hall-kit/master/install.sh | sh
```

- 加速前缀**内置**为 `https://gh-proxy.com/`，可用 `GH_PROXY` 覆盖：`… | GH_PROXY=https://your-mirror/ sh`
- 直连可用时优先直连，只有失败才自动走加速；标准代理 `https_proxy` 依然生效

`install.sh` 的环境变量：`HALL_KIT_VERSION`（release tag，默认 `latest`）、`HALL_KIT_INSTALL_DIR`（默认 `~/.hall/bin`）、`GH_PROXY`（覆盖内置加速前缀）；装完会自动执行一次 `hall shell install`。

> `latest` 通过 GitHub API 解析（匿名限流 60 次/小时）；限流或离线时用 `HALL_KIT_VERSION=v0.0.1` 指定版本。

或者用 Go（需要 Go >= 1.26）：

```bash
go install github.com/QuietSugar/hall-kit/cmd/hall@latest
hall aliases     # 在本二进制旁创建 gitcheck/gitclone/... 短名
```

> Windows：在 git-bash 中运行安装脚本（用 `tar` 解包）；Windows 支持为实验性，部分功能可能不可用。

### Windows：git-bash 与 Windows Terminal

安装脚本在 Windows 上把二进制放进 `%USERPROFILE%\.hall\bin`，并为每个短名生成两种 shim：

- `gitcheck`（POSIX shim）——git-bash / Cygwin 使用；
- `gitcheck.cmd`——cmd.exe / PowerShell（Windows Terminal）使用。

要在 Windows Terminal 里使用，先把该目录加入**用户 PATH**（本工具不会自动改 PATH）。PowerShell 执行一次：

```powershell
$dir = "$env:USERPROFILE\.hall\bin"
$p = [Environment]::GetEnvironmentVariable('Path','User')
if (($p -split ';') -notcontains $dir) { [Environment]::SetEnvironmentVariable('Path', ($p.TrimEnd(';') + ';' + $dir), 'User') }
```

重开 Windows Terminal 后：

```powershell
hall --help
hall gitcheck C:\path
gitcheck C:\path        # 短名，由 gitcheck.cmd 转发
```

`proxy` / `unproxy` / `tat` 是 bash 系专属；PowerShell 里请用 `$env:HTTPS_PROXY` 等环境变量。

## 使用方法

```bash
hall --help            # 查看用法
hall --version         # 查看版本信息
hall gitcheck ~/code   # 运行某个工具
```

短名（`dockerage`、`gitcheck`、`gitclone`、`gitarc`、`rmempty`）由 `hall aliases` 创建，指向同一个 `hall` 二进制（Unix 用相对软链，git-bash 用 shim 脚本），因此 `gitcheck ...` 与 `hall gitcheck ...` 完全等价。

> `cmd/<tool>` 仍保留为兼容 shim，单独 `go install .../cmd/gitcheck@latest` 也可以用。

## 工具列表

| 工具        | 说明                                       |
| ----------- | ------------------------------------------ |
| `hall`      | 工具箱总入口，打印项目信息和版本           |
| `dockerage` | 由 docker 镜像创建时间戳计算镜像年龄（秒） |
| `gitcheck`  | 检查 git 仓库是否可以安全删除（不丢失任何内容） |
| `gitclone`  | 按远程 URL 结构把仓库克隆到归档目录         |
| `gitarc`    | 把目录树中的 git 仓库按远程 URL 归档         |
| `rmempty`   | 删除目录树中的空目录（并清理 .DS_Store）     |


>
> 命令与 [hall-command](https://github.com/QuietSugar/hall-command) shell 版同名同源：`gitcheck`、`gitclone`、`gitarc`、`rmempty`。

每个工具也可以单独 `go install`（`cmd/<tool>` 是兼容 shim），例如：

```bash
go install github.com/QuietSugar/hall-kit/cmd/gitcheck@latest
```

### dockerage

```bash
dockerage "2025-11-07T11:12:09.744706645+08:00"   # 输出该镜像创建至今的秒数
```

时间戳可用 `docker image inspect --format '{{.Created}}' <镜像>` 获取。输出一个整数秒；非法时间或未来时间会报错并返回退出码 2。同样支持 `--lang zh|en` 或 `HALL_LANG`。

### gitcheck

`gitcheck` 递归扫描目标目录下的所有 git 仓库（默认当前目录），逐仓库检查未跟踪文件、未提交改动、stash、未推送提交、游离提交、远程配置、子模块、LFS 等，默认每个仓库输出一行汇总，`-v` 展示每项检查的耗时与未通过的详细原因。`--verify-remote` 模式会联网验证远程（单次调用 15 秒超时快速失败，网络不可达时提示代理配置方法）。退出码：`0` = 全部可删除，`1` = 存在不可删除的仓库，`2` = 参数/环境错误。

输出语言支持中文和英文，默认按 `HALL_LANG`、`LC_ALL`、`LANG` 环境变量自动探测，也可手动指定：

```bash
gitcheck --lang zh              # 中文输出
HALL_LANG=zh gitcheck ~/code    # 通过环境变量指定
```

### gitclone

```bash
gitclone https://github.com/QuietSugar/hall-kit.git
```

运行时打印远程地址和目标绝对路径；Windows 下额外打印同一路径在 git-bash（`/c/...`）或 Cygwin（`/cygdrive/c/...`）中的形式。目标已存在时提示跳过。

按 URL 结构把仓库克隆到归档目录，路径由 URL 推导（`https://github.com/user/repo.git` → `<归档目录>/github.com/user/repo`），目标已存在时跳过。归档目录取 `HALL_GIT_REPO_PATH`，未设置时读取 `~/.hall-command/env`（两种 shell 写法都支持），都没有则用 `~/git-repo`（优先级与 shell 版相反，见下文）。退出码：`0` = 成功，`1` = 运行失败，`2` = 参数/环境错误。

### gitarc

```bash
gitarc ~/code
```

递归扫描目录下最外层的 git 仓库，读取每个仓库的 `remote.origin.url` 并按 URL 结构移动到归档目录；地址无效或目标已存在时跳过，移动失败会继续处理其余仓库并最终返回退出码 1。归档目录解析规则同 `gitclone`。

### rmempty

```bash
rmempty ~/code   # 删除空目录，并清理 .DS_Store；.git 目录不动
```

递归删除目标目录（默认当前目录）下的空目录，先清理 `.DS_Store`，再从最深一层往上删除空目录；`.git` 及其内容始终保留。退出码：`0` = 成功，`1` = 运行失败，`2` = 参数/环境错误。

## 与 hall-command shell 版的差异

行为不保证逐字节一致，主要差异：

- `gitarc` 遇到没有 `remote.origin.url` 的仓库会跳过并继续，而不是被 `set -e` 中断整个流程；
- `gitarc` 支持跨文件系统移动（`rename` 失败时回退为复制+删除），并会解析符号链接形式的扫描目录；隐藏子目录会被跳过，与 shell 的 `*` 通配一致；
- `gitcheck` 的汇总中 clean 数量始终正确（shell 只在 `-a` 时计数）；参数/环境错误统一返回退出码 2，而不是 shell 的 1；
- 归档目录解析：环境变量 `HALL_GIT_REPO_PATH` 优先于 `~/.hall-command/env`（shell 是配置文件覆盖环境变量），且支持 `export`、引号、行尾注释等写法；
- URL 校验额外接受 `https://gh-proxy.com/` 前缀并去除首尾空白；含 `.`、`..` 路径段的地址会被拒绝，防止写到归档目录之外。

旧 shell 源码原样保留在 [legacy/](legacy/README.md)（仅参考，不参与构建与安装）。

## shell 资产（hall shell）

Go 工具之外，hall-kit 还内嵌一组可复用的 bash 资产（日志库、配置解析、`proxy`/`unproxy` 片段），由 `hall shell` 子命令管理，安装到 `~/.hall/`：

```bash
go install github.com/QuietSugar/hall-kit/cmd/hall@latest
hall shell install     # 解出资产到 ~/.hall/shell/，并幂等写入 ~/.bashrc/.zshrc/.profile
hall shell status      # 查看安装位置、文件与 rc 块状态
hall shell uninstall   # 删除资产与 rc 块（--purge 连 ~/.hall 一起删）
hall shell migrate     # 将旧 ~/.hall-command 配置导入 ~/.hall（可选）
```

- 配置：`~/.hall/env`（首次安装由模板生成，之后不会被覆盖）；存在旧 `~/.hall-command/env` 时回退读取
- 库：`~/.hall/shell/lib/`（`slog.sh`、`init.sh`、`load_config.sh`），由脚本显式 source
- 片段：`~/.hall/shell/source/`（内置的以 `hall-` 开头，如 `hall-env.sh`、`hall-proxy.sh`、`hall-docker.sh`），登录时自动 source；你自己的片段放同一目录、不加前缀，`install`/`uninstall` 不会碰
- 工具：`~/.hall/bin/`（`tat`：tmux 会话助手，需要本机安装 tmux），由同一 rc 块加入 `PATH`
- `--no-rc` 不动 rc 文件；`--dry-run` 只打印将要做的改动
- 资产内嵌在二进制里；`go install` 新版本后再运行一次 `hall shell install` 即为更新（当前不做联网下载）

旧 `~/.hall-command` 目录默认只迁移不删除；`hall shell migrate --purge` 会把它重命名为带时间戳的备份，用户文件不会丢。

## 网络访问与服务器引导

直连 GitHub 可用时，上面的 `go install` 就够了。访问不了 GitHub 时，Go 侧用内置的 `GOPROXY`：

```bash
go env -w GOPROXY=https://goproxy.cn,direct   # 一般只需设置一次
go install github.com/QuietSugar/hall-kit/cmd/hall@latest
hall shell install                            # 纯离线：资产内嵌在二进制里
```

shell 侧（[shell/lib/hall-net.sh](shell/lib/hall-net.sh)）统一约定：`https_proxy` 优先，其次 `GH_PROXY`（GitHub 加速前缀，如 `https://gh-proxy.com/`），默认直连。

如果是给一台新服务器做初始化，推荐顺序：

1. [ubuntu-server-init](https://github.com/QuietSugar/ubuntu-server-init) —— 机器引导层（Ubuntu/Debian 的镜像源、基础工具、docker、zsh，一次性）；
2. 再执行上面的 `go install` + `hall shell install` —— 工具箱层。

两层是**单向依赖**：引导层可以装工具箱，工具箱不依赖引导层。

## 新增一个工具

1. 逻辑放 `internal/<tool>/`，导出一个 `Main(args []string)`（参考 `internal/rmempty`）；
2. `cmd/<tool>/main.go` 留成 3 行 shim（兼容 `go install .../cmd/<tool>`）；
3. 在 `cmd/hall` 的 `toolTable` 和 `internal/toolalias.Names` 里登记；
4. 更新本 README 的工具表。

发布只构建 `cmd/hall`，短名由 `hall aliases` 生成。

## 用 gup 安装和管理本工具

[gup](https://github.com/golang-update/gup) 是 `go install` 的增强封装，可以先装 gup 本身：

```bash
go install github.com/golang-update/gup@latest
```

然后直接用 gup 安装本工具（效果等同于 `go install`，但安装前会自动备份旧版本，便于回滚）：

```bash
gup install github.com/QuietSugar/hall-kit/cmd/hall@latest
```

日常维护常用命令：

```bash
gup list       # 列出 gup 管理的所有命令及版本
gup update     # 批量把所有工具升级到最新版本
gup update hall  # 只升级 hall
```

## shell 小工具库

[shell/lib/slog.sh](shell/lib/slog.sh) 是可直接 source 的 bash 日志函数库（slog，MIT 许可），提供 `log_debug`、`log_info`、`log_success`、`log_warning`、`log_error`，输出级别由 `LOG_LEVEL_STDOUT`（默认 INFO）控制。它与 `legacy/` 中随旧脚本冻结的副本不同，会作为独立公共库长期保留，供任意 shell 脚本引用：

```bash
source /path/to/hall-kit/shell/lib/slog.sh
log_info "hello"
```

## 待办与路线图

- [ ] 清理 `hall shell migrate`：`~/.hall-command` 处理稳定后移除该临时命令。注意用户自建的 PATH 接线（例如 `~/.zsh/source/` 下的文件）需要手动处理，`hall shell` 不会改动个人文件。
- [ ] 远期：`hall shell install --from-github`（从网络更新 shell 资产）；当前 shell 资产仍是内嵌、离线，二进制已支持 release 安装。

## 许可证

本项目基于 [MIT License](LICENSE) 开源。
