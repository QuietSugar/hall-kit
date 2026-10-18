package i18n

// rmempty message templates, keyed by "rmempty.<area>.<name>".
var rmemptyMessages = map[Lang]map[string]string{
	ZH: {
		"rmempty.usage": `用法: rmempty [选项] [目录]

  递归删除目录（默认当前目录）下的空目录，并清理 .DS_Store 文件。
  .git 目录及其内容不会被删除。

参数:
  [目录]            要处理的根目录（默认当前目录）

选项:
  --lang <语言>     输出语言: en 或 zh（默认自动探测，也可通过 HALL_LANG 指定）
  -h, --help        显示帮助

退出码: 0 = 成功, 1 = 运行失败, 2 = 参数/环境错误`,

		"rmempty.err.unknownflag": "未知参数: %s",
		"rmempty.err.usage":       "错误: 只接受一个目录参数",
		"rmempty.err.abspath":     "错误: 无法解析路径: %s",
		"rmempty.err.nodir":       "错误: 目录不存在: %s",
		"rmempty.err.scan":        "处理目录失败: %v",
		"rmempty.info.target":     "此次操作目录: %s\n\n",
		"rmempty.removed":         "删除目录: %s",
		"rmempty.summary":         "完成：删除 %d 个空目录，%d 个 .DS_Store",
	},
	EN: {
		"rmempty.usage": `Usage: rmempty [options] [directory]

  Recursively remove empty directories under the target (default: current
  directory) and clean up .DS_Store files. The .git directory and its
  contents are never removed.

Arguments:
  [directory]         Root directory to process (default: current directory)

Options:
  --lang <lang>       Output language: en or zh (auto-detected by default,
                      or set via HALL_LANG)
  -h, --help          Show this help

Exit codes: 0 = success, 1 = runtime failure, 2 = argument/env error`,

		"rmempty.err.unknownflag": "unknown argument: %s",
		"rmempty.err.usage":       "error: only one directory argument is accepted",
		"rmempty.err.abspath":     "error: cannot resolve path: %s",
		"rmempty.err.nodir":       "error: directory does not exist: %s",
		"rmempty.err.scan":        "cannot process directory: %v",
		"rmempty.info.target":     "Processing: %s\n\n",
		"rmempty.removed":         "removed directory: %s",
		"rmempty.summary":         "Done: removed %d empty directories and %d .DS_Store files",
	},
}

func init() {
	register(rmemptyMessages)
}
