# 安装、升级与故障排查

中文 | [English](installation.en.md)

> 版本说明：下文“安装后直接启动”的持久化鉴权模式属于 v0.1.2 安装器（待发布）。已发布的 v0.1.1 仍使用旧环境变量模式；需要安装包含本次改动的新版本，或从新版源码构建后重新安装集成。现有 Release 不会被替换。

## 选择安装方式

普通用户优先使用 [GitHub Release](https://github.com/xsy-tech/erp-agent-skills/releases) 的安装包。安装 `erp-skill` 不需要 Go、GitHub CLI、仓库写权限或源码编译。它是 Agent 适配器安装与启动工具，不能代替 Codex、Claude Code 或 WorkBuddy 客户端。

准备好目标客户端、管理员提供的 MCP 地址，以及绑定到你本人 ERP 用户和租户的个人 API Key。新版安装器将 Codex / Claude Code 的鉴权写入当前用户的私有 MCP 配置，不再要求系统钥匙串或 Linux Secret Service。旧安装的系统凭证仅作为兼容读取来源。

| 操作系统 | 处理器 | v0.1.1 安装包 |
| --- | --- | --- |
| macOS | Intel | `erp-agent-skill_0.1.1_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `erp-agent-skill_0.1.1_darwin_arm64.tar.gz` |
| Linux | Intel / AMD 64 位 | `erp-agent-skill_0.1.1_linux_amd64.tar.gz` |
| Linux | ARM 64 位 | `erp-agent-skill_0.1.1_linux_arm64.tar.gz` |
| Windows | x64 | `erp-agent-skill_0.1.1_windows_amd64.zip` |

这些都是 `erp-skill` 工具包，**不是 WorkBuddy 可直接导入的 Connector**。新版 WorkBuddy 推荐直接安装 MCP + Skill；只有选择 Connector 分发方式时才需要生成目录。

## 一键安装已发布版本

### macOS / Linux

在 Bash / Zsh 终端执行：

```bash
set -o pipefail && curl -fsSL https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.sh | ERP_SKILL_RELEASE_BASE_URL=https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1 bash && export PATH="$HOME/.local/bin:$PATH" && erp-skill version
```

脚本自动识别系统和架构，下载对应安装包与 `checksums.txt`，校验通过后才安装。此固定版本脚本需要 `curl`、`tar`、`shasum` 和 Bash；Linux 的 `shasum` 通常随 Perl 提供。若系统只有 `sha256sum`，可按下文手动安装，或使用更新后的仓库安装脚本。

当前终端设置不影响以后新开的终端。可以在 `~/.zshrc` 或 `~/.bashrc` 中加入下面一行，再打开新终端：

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Windows x64

在 PowerShell 中执行：

```powershell
$ErrorActionPreference = 'Stop'; & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.ps1').Content)) -Version '0.1.1' -BaseUrl 'https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1'; $env:Path = "$HOME\.local\bin;$env:Path"; erp-skill version
```

脚本会下载 ZIP、核对 SHA-256，再安装到 `$HOME\.local\bin`。如需永久使用，可在 Windows“编辑账户的环境变量”中，将这个目录加入用户 `Path`，然后重开终端。

### 使用仓库中的新版安装脚本

下载或检出本仓库后，可执行以下命令。新版脚本默认使用官方 GitHub Release 地址；Shell 脚本会优先使用 `sha256sum`，其次使用 `shasum`：

```bash
bash scripts/install.sh
# 指定另一个已经发布的版本：
ERP_SKILL_VERSION=0.1.1 bash scripts/install.sh
# 使用组织自己的下载镜像和安装目录：
ERP_SKILL_VERSION=0.1.1 ERP_SKILL_RELEASE_BASE_URL=https://downloads.example.com/erp/v0.1.1 ERP_SKILL_INSTALL_DIR="$HOME/.local/bin" bash scripts/install.sh
```

```powershell
& .\scripts\install.ps1
& .\scripts\install.ps1 -Version '0.1.1' -BaseUrl 'https://downloads.example.com/erp/v0.1.1' -InstallDir "$HOME\.local\bin"
```

所有示例地址中的 `example.com` 都是占位符。版本号与镜像内安装包、校验清单必须一致。

## 手动安装预编译包

如果不方便运行一键脚本，可在 [v0.1.1 的 Assets](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1) 下载匹配平台的包和 `checksums.txt`。

macOS / Linux 先核对清单中对应文件的 SHA-256。macOS 示例：

```bash
shasum -a 256 erp-agent-skill_0.1.1_darwin_arm64.tar.gz
```

Linux 可运行 `sha256sum <安装包文件名>`。不要对只下载了一个平台包的目录直接执行整个 `checksums.txt` 的批量校验，因为清单还包含其他平台的文件。

确认哈希一致后解压，以下以 Apple Silicon 为例：

```bash
tar -xzf erp-agent-skill_0.1.1_darwin_arm64.tar.gz
mkdir -p "$HOME/.local/bin"
install -m 0755 erp-agent-skill_0.1.1_darwin_arm64/erp-skill "$HOME/.local/bin/erp-skill"
export PATH="$HOME/.local/bin:$PATH"
erp-skill version
```

Windows 使用 `Get-FileHash <ZIP文件名> -Algorithm SHA256` 核对，再用 `Expand-Archive` 解压，将包内的 `erp-skill.exe` 放入用户 PATH 中的目录。

## 连接 Codex 或 Claude Code

### macOS / Linux（Bash）

先将 `MCP_URL` 改成实际地址；Codex 使用 `AGENT=codex`，Claude Code 使用 `AGENT=claude-code`：

```bash
MCP_URL='https://erp-open.example.com/mcp'
AGENT=codex
IFS= read -r -s -p '请输入个人 ERP API Key：' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install "$AGENT" --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor "$AGENT" --url "$MCP_URL"
```

这里的 `read -p` 是 Bash 语法；如平常使用 Zsh，请先运行 `bash` 再粘贴这一段。

### Windows（PowerShell）

```powershell
$McpUrl = 'https://erp-open.example.com/mcp'
$Agent = 'codex' # Claude Code 改为 claude-code
$SecureKey = Read-Host '请输入个人 ERP API Key' -AsSecureString
$KeyPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureKey)
try {
    $PlainKey = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($KeyPointer)
    $PlainKey | erp-skill install $Agent --url $McpUrl --key-stdin
    if ($LASTEXITCODE -ne 0) { throw 'Agent 安装失败，请检查错误信息。' }
} finally {
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($KeyPointer)
    Remove-Variable PlainKey, SecureKey -ErrorAction SilentlyContinue
}
erp-skill doctor $Agent --url $McpUrl
```

以上过程使用隐藏输入和标准输入，不把 Key 写进命令参数或聊天。安装器会将 Key 保存到当前用户的 MCP 私有配置中，用于客户端直接连接；配置文件不要提交到仓库或共享。

### 启动新会话

安装后重启客户端，之后照常使用：

```bash
codex
# 或：
claude
```

也可以直接打开使用同一用户配置的本地桌面 Code 会话。正在运行的客户端需要重新连接 MCP 或重启，才能重新读取已保存的配置。无需给终端设置 ERP Key 环境变量，也无需让 Agent 嵌套启动交互进程。

## 直接运行 codex 或 claude

新版 `erp-skill install` 会安装共享 Skill，并将 URL 和鉴权请求头持久化到用户级配置：

| 客户端 | 配置位置 | 鉴权字段 |
| --- | --- | --- |
| Codex CLI / 同一配置下的本地 Codex 桌面 | `$CODEX_HOME/config.toml`，默认 `~/.codex/config.toml` | `mcp_servers.erp-open.http_headers.Authorization` |
| Claude Code CLI / 本地桌面 Code 会话 | `~/.claude.json`；设置 `CLAUDE_CONFIG_DIR` 时为该目录下 `.claude.json` | `mcpServers.erp-open.headers.Authorization` |

Codex 配置形式如下，示例中的 Key 是占位符，真实值由安装器从标准输入接收并写入：

```toml
[mcp_servers.erp-open]
url = "https://erp-open.example.com/mcp"
[mcp_servers.erp-open.http_headers]
Authorization = "Bearer <YOUR_PERSONAL_KEY>"
```

Claude Code 的对应形式：

```json
{
  "mcpServers": {
    "erp-open": {
      "type": "http",
      "url": "https://erp-open.example.com/mcp",
      "headers": { "Authorization": "Bearer <YOUR_PERSONAL_KEY>" }
    }
  }
}
```

这些是用户级私有文件，不写进项目 `.mcp.json` 或 Git 仓库；Unix 文件权限设为 `0600`，Windows 使用用户目录的访问控制。更新只修改 `erp-open`，保留其他服务与设置。安装器不会把 Key 交给客户端命令行参数，也不会在输出中展示它。检测到无效配置时会停止，避免覆盖现有文件。

同一账户下重启客户端后，MCP 可直接读取鉴权；Skill 可被发现，正文按相关请求按需加载。`initialize` 仅是 MCP 握手，本身不会安装 Skill。

Skill 当前安装到 `$CODEX_HOME/skills/erp-open-platform`（默认 `~/.codex/skills/erp-open-platform`）或 `~/.claude/skills/erp-open-platform`；设置 `CLAUDE_CONFIG_DIR` 时，Claude Skill 位于该目录的 `skills/erp-open-platform`。不同版本或禁用策略可能影响发现行为，未显示时应检查客户端设置。

`erp-skill launch codex` / `erp-skill launch claude-code` 保留为兼容入口。`doctor` 和 `launch` 优先读取持久化配置中的凭证，找不到时才尝试旧的系统凭证。新版安装无需系统凭证管理器。

**旧版本迁移：** v0.1.1 仍使用 `ERP_OPEN_CODEX_KEY` / `${ERP_OPEN_CLAUDE_KEY}` 和钥匙串，需要原有 `launch` 流程。安装包含此改动的新版本后，按安全输入命令重新执行 `install`；若 Skill 内容冲突，先备份并确认，再使用 `--force`。这会用固定请求头替换旧 ERP 环境变量配置。然后重启客户端。

**桌面范围：** 本节针对读取这些本机配置的 Codex 和 Claude Code 本地 Code 会话。普通 Claude Desktop 聊天界面的配置以及云端会话不同，不能据此保证已接入；其导入和授权应按对应客户端完成。

参考：[Codex MCP](https://developers.openai.com/codex/mcp/)、[Codex 配置参考](https://developers.openai.com/codex/config-reference/)、[Claude Code MCP](https://code.claude.com/docs/en/mcp)、[Claude Code 桌面](https://code.claude.com/docs/en/desktop)。

## 连接 WorkBuddy

### 默认：直接安装 MCP + Skill（v0.1.2）

本功能属于待发布的 v0.1.2；当前已发布的 v0.1.1 只有 Connector 导出功能。请使用包含此改动的新版安装器，或从新版源码构建后执行以下命令。

Bash（Zsh 用户先执行 `bash`），将地址换成实际端点：

```bash
MCP_URL='https://erp-open.example.com/mcp'
IFS= read -r -s -p '请输入个人 ERP API Key：' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install workbuddy --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor workbuddy --url "$MCP_URL"
```

Windows 使用上文 PowerShell 隐藏输入代码，将 `$Agent` 设为 `'workbuddy'` 即可；安装命令仍为 `erp-skill install $Agent --url $McpUrl --key-stdin`，之后执行 `erp-skill doctor workbuddy --url $McpUrl`。

安装器先进行只读 MCP 验证，再保存当前用户的 MCP 配置和共享 Skill：

| 内容 | 路径或字段 |
| --- | --- |
| 用户级 MCP 配置 | `~/.workbuddy/mcp.json`（Windows 为 `%USERPROFILE%\.workbuddy\mcp.json`） |
| ERP 条目 | `mcpServers.erp-open` |
| 传输类型 | `streamableHttp` |
| 凭证 | `headers.Authorization`，值为 `Bearer <个人Key>` |
| ERP Skill | `~/.workbuddy/skills/erp-open-platform/SKILL.md` 及中英文 references |

更新会保留其他 MCP、设置和 JSON 数值；无效配置会拒绝覆盖。Key 仅保存到用户级私有配置，不通过进程参数或聊天传递；Unix 权限为 `0600`。直接安装无需 Connector 包、Token 表单或 `erp-skill launch`。

完成后重启 WorkBuddy 或在“自定义连接器 / MCP”界面重连 `erp-open`，确认 ERP Skill 已启用。若当前项目存在同名项目级 MCP 配置或 Skill，需核对其优先级；安装器只处理用户级配置，不修改项目文件。

`doctor workbuddy --url ...` 默认读取已保存的 Key；保留 `--key-stdin` 作为独立检查方式，它只检查本次输入的 Key，不会覆盖已保存的凭证。

### 可选：导出 Connector 包

仅在需要分发带有 Token 表单、图标和 Skill 的 Connector 时使用：

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

该命令只生成目录，不写本机 WorkBuddy MCP、Key 或 Skill。它不需要 Key，不能同时传 `--key-stdin`。文件保留 `${ERP_OPEN_API_KEY}` 占位符；在 WorkBuddy 导入后由 Token 表单保存凭证。旧 v0.1.1 用户仍可使用这条路径。

### 从 Connector 迁移到直接安装

先在 WorkBuddy 中停用旧 ERP Connector，保留其文件；不要让旧 Connector 和新的 `erp-open` 同时连接，避免重复工具。使用新版安装器直接安装并验证，再重启 WorkBuddy、检查 MCP 和 Skill；确认新连接工作后，可在界面移除旧 Connector。安装器不会擅自删除已导入的 Connector。

[WorkBuddy 官方 MCP 指南](https://www.codebuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/MCP-Guide)说明用户级路径和配置方式。文件写入及协议验证与 WorkBuddy 实际界面验收是不同层次，安装器成功不表示已完成 UI 实测。

## 升级、回退与卸载

升级或回退：选择目标的已发布版本，使用对应脚本、安装包和校验清单重新安装 `erp-skill`，再检查 `erp-skill version`。

三个客户端直接安装的 Skill 内容变化时，安装器会拒绝覆盖内容不同的已有文件。先备份自己修改过的 Skill；确认要替换后，按上述安全输入方式，在 `install` 命令末尾添加 `--force`。只有使用可选 Connector 分发方式时，才为重新生成命令添加 `--force` 并重新导入。

三个客户端均可卸载直接安装的集成，例如：

```bash
erp-skill uninstall codex --url https://erp-open.example.com/mcp
# 或：
erp-skill uninstall claude-code --url https://erp-open.example.com/mcp
```

直接安装的 WorkBuddy 使用 `erp-skill uninstall workbuddy --url <实际MCP地址>`，移除本地 ERP MCP 凭证和未修改的托管 Skill，保留其他配置。可选 Connector 仍在 WorkBuddy 界面移除及清理 Token。确认所有客户端都不再使用工具后，可以手动移除 `~/.local/bin/erp-skill` 或 Windows 对应的 `erp-skill.exe`。本地卸载不会吊销服务端 API Key；丢失设备或怀疑凭证泄露时，应联系管理员吊销或轮换。

## 常见问题

| 现象 | 处理方法 |
| --- | --- |
| `erp-skill` 命令不存在 | 将安装目录加入 PATH，或使用可执行文件的完整路径。 |
| 下载 404 | 核对版本、平台、文件名和镜像地址；只有正式上传的 Assets 才可下载。 |
| SHA-256 校验失败 | 停止安装，重新下载匹配版本的包和清单；不要绕过校验。 |
| Linux 找不到 `shasum` | 固定的 v0.1.1 脚本需要该命令；安装 Perl、改用手动校验，或使用仓库新版脚本。 |
| 旧版凭证管理器不可用 | 旧 v0.1.1 可修复钥匙串；新版安装器使用用户级私有 MCP 配置，无此依赖。 |
| 提示会覆盖已修改的 Skill / Connector | 先备份并检查已有文件，确认后才使用 `--force`。 |
| `INSTALLED_AUTHORIZATION_REQUIRED` | 初始化和工具发现成功，但缺少只读的 `get_platform_info` 授权，请联系管理员。 |
| 直接启动客户端但调用失败 | 确认已用新版安装器重新安装并重启客户端，运行 doctor 检查实际结果，核对 Key 是否有效以及配置是否被项目级设置覆盖。 |
| Agent 看不到新增工具 | 重连 MCP 或正常重启客户端；以 `tools/list` 为准。 |
| WorkBuddy 只生成了目录 | 使用了可选的 --output 导出模式；要直接安装，应省略 --output，改用 --key-stdin。 |

另见[安装提示词](agent-install-prompts.md)、[Agent 使用说明](agent-usage.md)及[错误处理](../references/errors.md)。
