package i18n

// dockerage message templates, keyed by "dockerage.<area>.<name>".
var dockerageMessages = map[Lang]map[string]string{
	ZH: {
		"dockerage.usage": `用法: dockerage [选项] <RFC3339 时间戳>

  由 docker 镜像的创建时间戳计算镜像年龄（距当前时间的秒数），
  输出一个整数秒，便于脚本使用

参数:
  <RFC3339 时间戳>  例如 2025-11-07T11:12:09.744706645+08:00，
                    可用 docker image inspect --format '{{.Created}}' 获取

选项:
  --lang <语言>     输出语言: en 或 zh（默认自动探测，也可通过 HALL_LANG 指定）
  -h, --help        显示帮助

退出码: 0 = 成功, 2 = 参数/环境错误`,

		"dockerage.error.badtime":     "时间解析失败: %v",
		"dockerage.error.future":      "错误: 输入时间必须是过去的时间\n输入: %s\n当前: %s",
		"dockerage.error.toomany":     "错误: 只接受一个时间参数，多余参数: %s",
		"dockerage.error.unknownflag": "未知参数: %s",
	},
	EN: {
		"dockerage.usage": `Usage: dockerage [options] <RFC3339 timestamp>

  Print how many seconds have passed since a Docker image was created.
  A single integer is written to stdout, so the result is script-friendly.

Arguments:
  <RFC3339 timestamp>  e.g. 2025-11-07T11:12:09.744706645+08:00, as printed
                       by docker image inspect --format '{{.Created}}'

Options:
  --lang <lang>        Output language: en or zh (auto-detected by default,
                       or set via HALL_LANG)
  -h, --help           Show this help

Exit codes: 0 = success, 2 = argument/env error`,

		"dockerage.error.badtime":     "cannot parse timestamp: %v",
		"dockerage.error.future":      "error: the timestamp must be in the past\ninput: %s\nnow: %s",
		"dockerage.error.toomany":     "error: only one timestamp is accepted; extra argument: %s",
		"dockerage.error.unknownflag": "unknown argument: %s",
	},
}

func init() {
	register(dockerageMessages)
}
