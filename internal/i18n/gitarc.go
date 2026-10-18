package i18n

// gitarc message templates, keyed by "gitarc.<area>.<name>".
var gitarcMessages = map[Lang]map[string]string{
	ZH: {
		"gitarc.usage": `用法: gitarc [选项] [目录]

  递归扫描目录（默认当前目录）下最外层的 git 仓库，读取其
  remote.origin.url，并按 URL 结构移动到归档目录
  （HALL_GIT_REPO_PATH，默认 ~/git-repo）下。

参数:
  [目录]            要扫描的根目录（默认当前目录）

选项:
  --lang <语言>     输出语言: en 或 zh（默认自动探测，也可通过 HALL_LANG 指定）
  -h, --help        显示帮助

退出码: 0 = 成功, 1 = 有仓库移动失败, 2 = 参数/环境错误`,

		"gitarc.err.unknownflag": "未知参数: %s",
		"gitarc.err.nogit":       "错误: 找不到 git 命令",
		"gitarc.err.nodir":       "错误: 目录不存在: %s",
		"gitarc.err.abspath":     "错误: 无法解析路径: %s",
		"gitarc.err.root":        "无法确定或创建归档目录: %v",
		"gitarc.err.scan":        "扫描目录失败: %v",
		"gitarc.err.mkdir":       "无法创建目录 %s: %v",
		"gitarc.err.move":        "移动失败 %s -> %s: %v",
		"gitarc.info.scan":       "此次操作目录: %s\n归档目录:   %s\n\n",
		"gitarc.info.move":       "移动: %s -> %s",
		"gitarc.warn.noremote":   "跳过（没有配置远程）: %s",
		"gitarc.warn.invalid":    "跳过（remote.origin.url 不是支持的 Git 地址 %s）: %s",
		"gitarc.warn.exists":     "目标已存在,跳过: %s",
		"gitarc.scan.norepos":    "未找到 Git 仓库",
		"gitarc.scan.summary":    "完成：移动 %d，跳过 %d，失败 %d，共 %d 个仓库",
	},
	EN: {
		"gitarc.usage": `Usage: gitarc [options] [directory]

  Recursively scan the directory (default: current directory) for the
  outermost git repositories, read each remote.origin.url and move the
  repository under the archive root (HALL_GIT_REPO_PATH, default
  ~/git-repo) following the URL layout.

Arguments:
  [directory]         Root directory to scan (default: current directory)

Options:
  --lang <lang>       Output language: en or zh (auto-detected by default,
                      or set via HALL_LANG)
  -h, --help          Show this help

Exit codes: 0 = success, 1 = some repositories could not be moved,
2 = argument/env error`,

		"gitarc.err.unknownflag": "unknown argument: %s",
		"gitarc.err.nogit":       "error: git command not found",
		"gitarc.err.nodir":       "error: directory does not exist: %s",
		"gitarc.err.abspath":     "error: cannot resolve path: %s",
		"gitarc.err.root":        "cannot determine or create the archive directory: %v",
		"gitarc.err.scan":        "cannot scan directory: %v",
		"gitarc.err.mkdir":       "cannot create directory %s: %v",
		"gitarc.err.move":        "cannot move %s -> %s: %v",
		"gitarc.info.scan":       "Scanning: %s\nArchive:  %s\n\n",
		"gitarc.info.move":       "moved: %s -> %s",
		"gitarc.warn.noremote":   "skipping (no remote configured): %s",
		"gitarc.warn.invalid":    "skipping (unsupported remote.origin.url %s): %s",
		"gitarc.warn.exists":     "target already exists, skipping: %s",
		"gitarc.scan.norepos":    "No git repositories found",
		"gitarc.scan.summary":    "Done: %d moved, %d skipped, %d failed, %d repos total",
	},
}

func init() {
	register(gitarcMessages)
}
