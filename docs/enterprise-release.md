# 构建、企业分发与 GitHub 发布

本文面向维护者。普通用户请直接使用[预编译包安装指南](installation.md)，无需安装 Go 或执行构建。

## 仓库与权限

默认发布到 GitHub 的 `xsy-tech/erp-agent-skills`。仓库写入和 Tag 权限应由维护者持有，普通用户只需下载版本固定的安装包。

## 本地开发与构建（备选安装方式）

修改源码时，使用 `go.mod` 声明的 Go 版本：

```bash
git clone https://github.com/xsy-tech/erp-agent-skills.git
cd erp-agent-skills
make check
make build
./bin/erp-skill version
```

`make build` 生成本机工具；`make dist` 交叉编译五个平台并生成 `dist/checksums.txt`。`make help` 列出所有入口。

## 一条命令发布

```bash
make release
# 或明确核对将要发布的源码版本：
make release VERSION=0.1.1
```

发布前需要：

1. 为新版本同步更新 `version.go`、`manifest.json`、`adapters/workbuddy/connector-meta.json`、`scripts/install.sh` 和 `scripts/install.ps1`。同步更新 README、安装指南与提示词中的固定版本和下载地址。
2. 执行 `make release-check`，它依次运行测试、`go vet`、五个平台打包和 SHA-256 校验，不会创建 Tag 或 Release。
3. 检查源码与制品，确认没有凭证；提交并推送全部发布改动。工作区必须干净，包括未跟踪文件。
4. 安装 GitHub CLI，并使用有仓库写权限的账号执行一次 `gh auth login`。

`make release` 会再次运行检查，核实当前提交已经存在于 GitHub，并拒绝已有的版本 Tag。它先向草稿 Release 上传全部文件，全部成功后公开 Release；GitHub 将 `v<版本>` 指向已检查的提交，无需手动执行 `git tag`。

`VERSION` 只用于核对版本，不会替你修改源码版本。目标仓库可以显式配置：

```bash
make release REPO=owner/repository
```

## 发布内容与恢复

每个版本包含：

- macOS amd64 / arm64 两个 `.tar.gz`。
- Linux amd64 / arm64 两个 `.tar.gz`。
- Windows amd64 一个 `.zip`。
- 一个 `checksums.txt`。

版本包含 `-` 时会标记为预发布。上传失败时先检查 Releases 页面是否存在不完整草稿；公开失败时，已上传文件会保留在草稿中，可人工核对后继续发布。命令不会自动删除草稿、覆盖已有 Tag 或替换资产。

下载地址格式为：

```text
https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1/erp-agent-skill_0.1.1_darwin_arm64.tar.gz
```

**已发布包保持不可变。** v0.1.1 内置的 README、Skill 和 Connector 模板以当时的发布提交为准。仓库后续的中文文档与模板改动，需要在新版本构建发布后才会进入下载包，不要直接覆盖 v0.1.1 的资产。发布新版本后，核对一键安装命令实际使用的脚本 Tag、包版本和校验文件一致。

先向小范围用户试用，再按固定版本逐步推广。

## 文档维护约定

所有面向使用者的说明以中文提供，包括 README、`docs/`、共享 `SKILL.md`、`references/`、安装脚本提示和 WorkBuddy 显示字段。命令、API 字段、错误码和环境变量保留原名。新增文档应从 README 导航可达。

安装说明优先给出 Release 包和一键安装命令，其次提供手动下载，最后才是源码构建。复制到 Agent 的提示词应说明版本、平台识别、校验、安全输入、客户端重连及真实完成条件。

## 个人 API Key 与审计

按用户和环境分别签发 Key，绑定正确的 ERP 用户、租户与最小必需能力，不要多人共用或将 Key 打进 Connector 包。

Codex / Claude Code 使用操作系统凭证管理器：macOS Keychain、Windows 凭证管理器、Linux Secret Service。无桌面的 Linux 也需要可用后端；失败时应给出可操作错误，不能回退到明文文件。WorkBuddy 用户在客户端 Token 界面填写 Key。

开放平台记录 Client、API Key 标识、绑定用户、租户、Capability、协议、请求 ID、结果与耗时。发布维护者保留制品校验值、Tag/提交、推广分组和回退记录；日志和支持材料不得包含 Key 明文。

## 回退与凭证事件

回退到上一个已验证的版本包，并重新安装相应集成。回退 Skill 不会改变服务端能力授权；个人 Key 仍由系统凭证管理器或客户端 Token 界面管理。

凭证疑似泄露时先在开放平台吊销 Key，检查审计记录，再按需要签发新 Key。本地卸载不能代替服务端吊销。
