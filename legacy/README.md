# legacy/ — 旧 shell 实现的参考副本

这里的脚本原本位于 [hall-command](https://github.com/QuietSugar/hall-command)，
对应功能已用 Go 重写并迁移到本仓库：

| shell | Go 实现 |
| --- | --- |
| gitcheck.sh | cmd/gitcheck |
| gitclone.sh | cmd/gitclone |
| gitarc.sh | cmd/gitarc |
| lib/git_tool.sh | internal/giturl |
| lib/load_config.sh | internal/gitenv |

## 状态

- **仅作参考，不参与构建、安装或运行**，也不会进入 PATH。
- 只接受故障修复，不再新增功能；已知 bug 已就地修复（见下）。
- 过渡期结束后整个 legacy/ 目录会被删除。

## 与原始版本的差异（故障修复）

- gitarc.sh：某个仓库没有 remote.origin.url 时不再因 set -e 中断整个流程；推导出的路径含 ../ 时跳过，避免移动到归档目录之外。
- gitcheck.sh：即使没有 -a，也会统计 CLEAN 数量，汇总信息正确。
- gitclone.sh：以 git clone 的退出状态判断成功与否，不再用 cd 的结果。

运行方式（仅调试参考，非支持路径）：

    cd legacy && ./gitcheck.sh -a <dir>

依赖 bash；lib/ 下的 init.sh 会读取同目录的 slog.sh 和 load_config.sh。
