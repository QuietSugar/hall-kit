package i18n

// gitcheck message templates, keyed by "gitcheck.<area>.<name>".
var gitcheckMessages = map[Lang]map[string]string{
	ZH: {
		"gitcheck.usage": `用法: gitcheck [选项] [目标路径]   （默认当前目录）

  递归扫描目标目录下所有 git 仓库（目标本身是仓库时只检查它），
  默认每个仓库一行汇总，-v 可展开详细原因

选项:
  -a               同时打印 CLEAN 的仓库
  -q               安静模式，不打印汇总信息
  -v               打印每个仓库未通过检查的详细原因
  --lang <语言>    输出语言：en 或 zh（默认自动探测，也可通过 HALL_LANG 指定）
  --strict         严格模式：被 gitignore 忽略的文件也判定为不可删除
  --verify-remote  联网验证远程仓库存在且包含本地所有分支/tag/LFS 对象
  -h, --help       显示帮助

退出码: 0 = 全部可删除, 1 = 存在不可删除的仓库, 2 = 参数/环境错误`,

		"gitcheck.error.unknownflag": "未知参数: %s",
		"gitcheck.error.nogit":       "错误: 找不到 git 命令",
		"gitcheck.error.badpath":     "错误: 无效路径: %s",
		"gitcheck.error.nodir":       "错误: 目录不存在: %s",
		"gitcheck.error.gitstatus":   "无法执行 git status: %v",

		"gitcheck.check.untracked.name":   "未跟踪/被忽略文件",
		"gitcheck.check.uncommitted.name": "未提交的改动",
		"gitcheck.check.stash.name":       "stash",
		"gitcheck.check.branches.name":    "分支推送状态",
		"gitcheck.check.dangling.name":    "游离提交",
		"gitcheck.check.tags.name":        "tag 推送状态",
		"gitcheck.check.remote.name":      "远程仓库",
		"gitcheck.check.remotealive.name": "远程可达性（联网）",
		"gitcheck.check.submodules.name":  "子模块",
		"gitcheck.check.lfs.name":         "Git LFS",

		"gitcheck.untracked.tag":   "[未跟踪%d]",
		"gitcheck.uncommitted.tag": "[未提交%d]",
		"gitcheck.stash.tag":       "[储藏%d]",

		"gitcheck.branches.noremote": "分支 %s: 仓库没有任何远程分支，%d 个提交未推送",
		"gitcheck.branches.unpushed": "分支 %s: %d 个提交未推送到任何远程分支",
		"gitcheck.branches.tag":      "[未推送%d]",

		"gitcheck.dangling.detail": "处于 detached HEAD 状态（%s），有 %d 个提交不属于任何分支或远程",
		"gitcheck.dangling.tag":    "[游离%d]",

		"gitcheck.tags.remoteerr": "无法访问远程 %s 验证 tag: %v",
		"gitcheck.tags.missing":   "tag %s 不存在于远程 %s",
		"gitcheck.tags.tag":       "[tag未推送%d]",

		"gitcheck.remote.none": "没有配置任何远程仓库，本地提交无法保证已备份",
		"gitcheck.remote.tag":  "[无远程]",

		"gitcheck.remotealive.unreachable": "远程 %s 不可达（仓库可能已被删除/迁移）: %v",
		"gitcheck.remotealive.tag":         "[远程不可达]",

		"gitcheck.submodules.dirty":    "%d 个未提交/未跟踪文件",
		"gitcheck.submodules.noremote": "没有配置远程",
		"gitcheck.submodules.unpushed": "%d 个提交未推送",
		"gitcheck.submodules.detail":   "子模块 %s: %s",
		"gitcheck.submodules.joiner":   "，",
		"gitcheck.submodules.tag":      "[子模块%d]",

		"gitcheck.lfs.offlinehint": "仓库使用 Git LFS，离线模式无法确认大对象是否已上传到远程，可用 --verify-remote 检查",
		"gitcheck.lfs.notool":      "仓库使用 Git LFS，但未安装 git-lfs，无法验证 LFS 对象是否已上传",
		"gitcheck.lfs.notool.tag":  "[LFS未验证]",
		"gitcheck.lfs.verifyerr":   "无法对远程 %s 验证 LFS 对象: %v",
		"gitcheck.lfs.pending":     "%d 个 LFS 对象未上传到远程 %s",
		"gitcheck.lfs.tag":         "[LFS未上传]",

		"gitcheck.scan.target":     "此次检查目录: %s\n\n",
		"gitcheck.scan.norepos":    "未找到 Git 仓库",
		"gitcheck.scan.col.status": "STATUS",
		"gitcheck.scan.col.branch": "BRANCH",
		"gitcheck.scan.col.path":   "PATH",
		"gitcheck.scan.dirty":      "[DIRTY]",
		"gitcheck.scan.clean":      "[CLEAN]",
		"gitcheck.scan.summary":    "检查完成：dirty=%d, clean=%d, 共 %d 个仓库",
		"gitcheck.scan.ok":         "[OK]",
		"gitcheck.scan.fail":       "[FAIL]",

		"gitcheck.net.timeout":     "访问远程 %s 超时（%d 秒内无响应）",
		"gitcheck.net.unreachable": "远程 %s 网络不可达: %v",
		"gitcheck.net.proxyhint":   "提示：如需代理访问，可为 HTTPS 远程配置 git config --global http.<host>.proxy http://127.0.0.1:7890，或为 SSH 远程配置 ~/.ssh/config 的 ProxyCommand",
	},
	EN: {
		"gitcheck.usage": `Usage: gitcheck [options] [target path]   (default: current directory)

  Recursively scan all git repositories under the target (the target itself
  is checked alone when it is a repo). Prints one summary line per repo
  by default; use -v to show detailed reasons.

Options:
  -a               Also print CLEAN repositories
  -q               Quiet mode: no summary line
  -v               Print detailed reasons for each failed check
  --lang <lang>    Output language: en or zh (auto-detected by default,
                   or set via HALL_LANG)
  --strict         Strict mode: files ignored by gitignore also count
  --verify-remote  Verify online that remotes exist and contain all
                   local branches/tags/LFS objects
  -h, --help       Show help

Exit codes: 0 = all safe to delete, 1 = some repos not safe, 2 = argument/env error`,

		"gitcheck.error.unknownflag": "unknown argument: %s",
		"gitcheck.error.nogit":       "error: git command not found",
		"gitcheck.error.badpath":     "error: invalid path: %s",
		"gitcheck.error.nodir":       "error: directory does not exist: %s",
		"gitcheck.error.gitstatus":   "cannot run git status: %v",

		"gitcheck.check.untracked.name":   "Untracked/ignored files",
		"gitcheck.check.uncommitted.name": "Uncommitted changes",
		"gitcheck.check.stash.name":       "stash",
		"gitcheck.check.branches.name":    "Branch push status",
		"gitcheck.check.dangling.name":    "Dangling commits",
		"gitcheck.check.tags.name":        "Tag push status",
		"gitcheck.check.remote.name":      "Remote repository",
		"gitcheck.check.remotealive.name": "Remote reachability (online)",
		"gitcheck.check.submodules.name":  "Submodules",
		"gitcheck.check.lfs.name":         "Git LFS",

		"gitcheck.untracked.tag":   "[untracked %d]",
		"gitcheck.uncommitted.tag": "[uncommitted %d]",
		"gitcheck.stash.tag":       "[stashed %d]",

		"gitcheck.branches.noremote": "branch %s: repo has no remote branches, %d commits unpushed",
		"gitcheck.branches.unpushed": "branch %s: %d commits not pushed to any remote branch",
		"gitcheck.branches.tag":      "[unpushed %d]",

		"gitcheck.dangling.detail": "detached HEAD (%s) with %d commits not on any branch or remote",
		"gitcheck.dangling.tag":    "[dangling %d]",

		"gitcheck.tags.remoteerr": "cannot reach remote %s to verify tags: %v",
		"gitcheck.tags.missing":   "tag %s missing on remote %s",
		"gitcheck.tags.tag":       "[tags unpushed %d]",

		"gitcheck.remote.none": "no remote configured; local commits are not backed up",
		"gitcheck.remote.tag":  "[no remote]",

		"gitcheck.remotealive.unreachable": "remote %s unreachable (repo may be deleted or moved): %v",
		"gitcheck.remotealive.tag":         "[remote unreachable]",

		"gitcheck.submodules.dirty":    "%d uncommitted/untracked files",
		"gitcheck.submodules.noremote": "no remote configured",
		"gitcheck.submodules.unpushed": "%d commits unpushed",
		"gitcheck.submodules.detail":   "submodule %s: %s",
		"gitcheck.submodules.joiner":   ", ",
		"gitcheck.submodules.tag":      "[submodules %d]",

		"gitcheck.lfs.offlinehint": "repo uses Git LFS; offline mode cannot confirm large objects are uploaded, use --verify-remote",
		"gitcheck.lfs.notool":      "repo uses Git LFS but git-lfs is not installed; cannot verify LFS objects",
		"gitcheck.lfs.notool.tag":  "[LFS unverified]",
		"gitcheck.lfs.verifyerr":   "cannot verify LFS objects on remote %s: %v",
		"gitcheck.lfs.pending":     "%d LFS objects not uploaded to remote %s",
		"gitcheck.lfs.tag":         "[LFS unsynced]",

		"gitcheck.scan.target":     "Scanning: %s\n\n",
		"gitcheck.scan.norepos":    "No git repositories found",
		"gitcheck.scan.col.status": "STATUS",
		"gitcheck.scan.col.branch": "BRANCH",
		"gitcheck.scan.col.path":   "PATH",
		"gitcheck.scan.dirty":      "[DIRTY]",
		"gitcheck.scan.clean":      "[CLEAN]",
		"gitcheck.scan.summary":    "Done: dirty=%d, clean=%d, %d repos total",
		"gitcheck.scan.ok":         "[OK]",
		"gitcheck.scan.fail":       "[FAIL]",

		"gitcheck.net.timeout":     "remote %s timed out (no response within %ds)",
		"gitcheck.net.unreachable": "remote %s network unreachable: %v",
		"gitcheck.net.proxyhint":   "hint: to access via a proxy, run git config --global http.<host>.proxy http://127.0.0.1:7890 for HTTPS remotes, or configure a ProxyCommand in ~/.ssh/config for SSH remotes",
	},
}

func init() {
	register(gitcheckMessages)
}
