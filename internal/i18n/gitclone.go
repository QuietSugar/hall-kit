package i18n

// gitclone message templates, keyed by "gitclone.<area>.<name>".
var gitcloneMessages = map[Lang]map[string]string{
	ZH: {
		"gitclone.usage": `用法: gitclone [选项] <仓库URL>

  按远程 URL 的结构，把仓库克隆到归档目录（HALL_GIT_REPO_PATH，
  默认 ~/git-repo）下。

参数:
  <仓库URL>         支持 https:// 与 git@host:path 两种形式

选项:
  --lang <语言>     输出语言: en 或 zh（默认自动探测，也可通过 HALL_LANG 指定）
  -h, --help        显示帮助

退出码: 0 = 成功, 1 = 运行失败, 2 = 参数/环境错误`,

		"gitclone.err.usage":       "错误: 需要且只需要一个仓库 URL",
		"gitclone.err.unknownflag": "未知参数: %s",
		"gitclone.err.invalid":     "无效的 Git 地址: %s",
		"gitclone.err.nogit":       "错误: 找不到 git 命令",
		"gitclone.err.root":        "无法确定或创建归档目录: %v",
		"gitclone.err.mkdir":       "无法创建目录 %s: %v",
		"gitclone.err.clone":       "git clone 失败: %v",
		"gitclone.info.url":        "【项目远程地址】%s",
		"gitclone.info.target":     "【项目绝对路径】%s",
		"gitclone.info.posix":      "【%s 路径】%s",
		"gitclone.warn.exists":     "已存在,跳过: %s",
	},
	EN: {
		"gitclone.usage": `Usage: gitclone [options] <repo-url>

  Clone a repository into the archive root (HALL_GIT_REPO_PATH, default
  ~/git-repo), mirroring the directory layout of the remote URL.

Arguments:
  <repo-url>          https:// or git@host:path form

Options:
  --lang <lang>       Output language: en or zh (auto-detected by default,
                      or set via HALL_LANG)
  -h, --help          Show this help

Exit codes: 0 = success, 1 = runtime failure, 2 = argument/env error`,

		"gitclone.err.usage":       "error: exactly one repository URL is required",
		"gitclone.err.unknownflag": "unknown argument: %s",
		"gitclone.err.invalid":     "invalid git URL: %s",
		"gitclone.err.nogit":       "error: git command not found",
		"gitclone.err.root":        "cannot determine or create the archive directory: %v",
		"gitclone.err.mkdir":       "cannot create directory %s: %v",
		"gitclone.err.clone":       "git clone failed: %v",
		"gitclone.info.url":        "[remote] %s",
		"gitclone.info.target":     "[target] %s",
		"gitclone.info.posix":      "[%s path] %s",
		"gitclone.warn.exists":     "already exists, skipping: %s",
	},
}

func init() {
	register(gitcloneMessages)
}
