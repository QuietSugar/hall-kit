# AGENTS.md

本文件为在本仓库工作的 AI 编码代理（以及人类贡献者）提供硬性规则。

## 仓库性质：公开项目

`github.com/QuietSugar/hall-kit` 是**公开**仓库。提交进入的内容会永久留在公开的 git 历史中，事后删除文件也无法真正抹去。

## 架构约定

- 发布的是**一个**多调用二进制 `hall`：`cmd/hall` 用 `toolTable` 把工具名映射到 `internal/<tool>.Main`，并按 `argv[0]` 分发（Unix 软链 / git-bash shim）。
- 各工具逻辑放 `internal/<tool>`；`cmd/<tool>` 只保留兼容 shim，别在里面写逻辑。
- 新增工具要同时登记到 `cmd/hall` 的 `toolTable` 和 `internal/toolalias.Names`，并更新 README 工具表。

## 发布与 tag 纪律

- **非必要不打 tag**：打 tag 会触发 GitHub Actions（`ci` + `release`），耗时较长。
- 需要发版时，**只递增第三位**（patch），例如 `v0.0.1` → `v0.0.2`。
- **递增第一/第二位**（minor / major，如 `v0.3.0`、`v1.0.0`）必须先得到仓库所有者明确同意。
- 多个小修复先攒批，一次只打一个 patch tag；纯文档、注释、CI 调整不单独打 tag。

## 绝对禁止提交任何敏感信息

包括但不限于：

- **凭据与密钥**：API token、access key、密码、PIN、SSH/GPG 私钥、证书私钥、session cookie、OAuth/JWT secret；
- **服务配置**：`.env`、`.env.*`、凭据文件、`*.pem`、`*.key`、`kubeconfig`、`docker/config.json`、云厂商 CLI 凭据；
- **私有信息**：内网地址与主机名、未公开的 URL、真实姓名/手机号/邮箱/身份证号等个人数据、客户或第三方未公开信息、内部工单链接；
- **个人信息**：包含用户名的绝对路径（如 `/home/<用户名>/...`）、公司内部仓库路径；
- **大文件与二进制**：构建产物、压缩包、模型权重（除非确有必要且明确获批）。

文档、注释、测试和提交信息中的示例配置只能使用 `127.0.0.1`、`example.com`、`<your-token>` 这类占位值。

## 提交前必做

1. 用 `git status --short` 和 `git diff --cached` 逐项检查暂存内容；
2. 快速扫描可疑字符串（辅助手段，人工确认仍不可少）：

   ```bash
   git diff --cached | grep -nEi 'token|secret|password|passwd|api[_-]?key|BEGIN .*PRIVATE KEY|AKIA[0-9A-Z]{16}'
   ```

3. 只 `git add` 明确要提交的文件，避免 `git add -A` 误带本地配置；
4. 新增依赖或配置文件前，先确认其中不含环境相关的秘密。

## 敏感信息进入工作区之后

- **尚未提交**：先取消暂存（`git restore --staged <file>`），把真实值改为占位符或移入环境变量；
- **已提交但未推送**：用 `git commit --amend` 或交互式 rebase 改写本地历史；
- **已经推送**：立即按泄露处理——吊销并轮换对应凭据，再联系仓库所有者清理历史；只删除文件是不够的；
- 不确定是否属于敏感信息时，停下并询问仓库所有者，不要继续推送。
