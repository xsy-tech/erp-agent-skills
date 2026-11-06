# ERP Agent Skill（中文使用指南）

让 Codex、Claude Code 和 WorkBuddy 通过 ERP 开放平台 MCP，使用你获授权的 ERP 业务能力。

**推荐直接安装 GitHub Release 的预编译包，不需要安装 Go、克隆仓库或编译源码。** 当前已发布版本为 [v0.1.1](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1)。

## 一键安装命令

以下命令只下载、校验和安装 `erp-skill`。首次使用还需按下文连接你的 Agent。请先安装目标 Agent，并向管理员获取 MCP 地址和个人 API Key。

### macOS / Linux（Bash 或 Zsh）

复制整行执行，自动选择 Intel / AMD 或 ARM 安装包，校验 SHA-256，安装到 `~/.local/bin`：

```bash
set -o pipefail && curl -fsSL https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.sh | ERP_SKILL_RELEASE_BASE_URL=https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1 bash && export PATH="$HOME/.local/bin:$PATH" && erp-skill version
```

Linux 运行此固定版本安装脚本需要 `shasum`；通常由 `perl` 包提供。仓库新版脚本同时支持 `sha256sum`，详见[安装指南](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.md)。

### Windows x64（PowerShell）

```powershell
$ErrorActionPreference = 'Stop'; & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.ps1').Content)) -Version '0.1.1' -BaseUrl 'https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1'; $env:Path = "$HOME\.local\bin;$env:Path"; erp-skill version
```

默认安装到当前用户目录，无需管理员权限。这两条命令只更新当前终端的 PATH；永久设置、手动下载、升级和故障排查见[完整安装指南](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.md)。

## 让 Agent 帮你安装

复制[安装提示词](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-install-prompts.md)中对应的一段，发送给 **Codex、Claude Code 或 WorkBuddy**。提示词会引导 Agent 优先安装预编译包，并检查版本、配置和验证结果。

**不要把个人 API Key 粘贴到聊天中。** Codex / Claude Code 在本机终端隐藏输入；WorkBuddy 在 Token 界面填写。

## 连接 Codex / Claude Code

以下示例适用于 macOS / Linux 的 Bash。将地址替换为管理员提供的真实 MCP 地址；API Key 会隐藏输入并通过标准输入传给安装器，不出现在命令参数中。

```bash
MCP_URL='https://erp-open.example.com/mcp'
AGENT=codex  # Claude Code 请改成 claude-code
IFS= read -r -s -p '请输入个人 ERP API Key：' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install "$AGENT" --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor "$AGENT" --url "$MCP_URL"
```

安装成功后退出当前 Agent 会话，从终端启动新会话：

```bash
erp-skill launch codex
# 或：erp-skill launch claude-code
```

安装器会安装共享 Skill、注册名为 `erp-open` 的 MCP 服务，并将个人凭证保存到系统凭证管理器。每次使用 `launch`，凭证才会注入所启动的 Agent 进程。Windows 的安全输入命令见[安装指南](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.md#连接-codex-或-claude-code)。

## 连接 WorkBuddy

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

在 WorkBuddy 的 Connector 导入入口选择生成的目录，并在 **Token 输入界面**填写个人 API Key，然后连接或重连 MCP。生成的文件只包含 `${ERP_OPEN_API_KEY}` 占位符。`doctor workbuddy` 是可选的独立连通性检查，不会替你保存 Token 或完成界面导入。

## 如何使用

连接后，可以用中文描述业务目标：

- “搜索姓名包含张学的员工，列出工号和部门。”
- “先查询这个订单当前状态，再告诉我是否可以提交。”
- “创建销售单前，先展示客户、商品、数量和金额，等我确认后再执行。”

可用工具以当前 MCP 会话的 `tools/list` 为准。Key 绑定的 ERP 用户、租户和数据范围由服务端决定；权限变更后重连 MCP 或启动新会话即可。

`doctor` 返回 `VERIFIED` 表示验证成功；`INSTALLED_AUTHORIZATION_REQUIRED` 表示已完成初始化和工具发现，但缺少 `get_platform_info` 授权，需要联系管理员。验证不会调用具有写入能力的业务工具。

## 中文文档导航

- [安装、升级、卸载与故障排查](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.md)
- [Codex / Claude Code / WorkBuddy 安装提示词](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-install-prompts.md)
- [Agent 使用说明](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-usage.md)
- [维护者构建与发布说明](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/enterprise-release.md)
- [共享 Skill 指令](https://github.com/xsy-tech/erp-agent-skills/blob/main/SKILL.md)
- [ERP 术语](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/terminology.md)、[业务工作流](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/workflows.md)、[安全规则](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/safety.md)、[错误处理](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/errors.md)

## 开发者：从源码构建（备选）

仅在需要修改或调试本项目时使用。Go 版本以 `go.mod` 为准：

```bash
git clone https://github.com/xsy-tech/erp-agent-skills.git
cd erp-agent-skills
make check
make build
./bin/erp-skill version
```

维护者使用 `make dist` 生成五个平台包，使用 `make release` 发布 GitHub Release，详见[发布说明](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/enterprise-release.md)。
