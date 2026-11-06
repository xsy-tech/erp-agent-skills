# 安装、升级与故障排查

中文 | [English](installation.en.md)

## 选择安装方式

普通用户优先使用 [GitHub Release](https://github.com/xsy-tech/erp-agent-skills/releases) 的安装包。安装 `erp-skill` 不需要 Go、GitHub CLI、仓库写权限或源码编译。它是 Agent 适配器安装与启动工具，不能代替 Codex、Claude Code 或 WorkBuddy 客户端。

准备好目标客户端、管理员提供的 MCP 地址，以及绑定到你本人 ERP 用户和租户的个人 API Key。Codex / Claude Code 使用系统凭证管理器：macOS Keychain、Windows 凭证管理器、Linux Secret Service。无桌面的 Linux 也必须有可用的 Secret Service 后端，缺失时安装会明确失败。

| 操作系统 | 处理器 | v0.1.1 安装包 |
| --- | --- | --- |
| macOS | Intel | `erp-agent-skill_0.1.1_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `erp-agent-skill_0.1.1_darwin_arm64.tar.gz` |
| Linux | Intel / AMD 64 位 | `erp-agent-skill_0.1.1_linux_amd64.tar.gz` |
| Linux | ARM 64 位 | `erp-agent-skill_0.1.1_linux_arm64.tar.gz` |
| Windows | x64 | `erp-agent-skill_0.1.1_windows_amd64.zip` |

这些都是 `erp-skill` 工具包，**不是 WorkBuddy 可直接导入的 Connector**。WorkBuddy 还需运行下文的生成命令。

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

以上过程使用隐藏输入和标准输入，不把 Key 字面值写进命令参数、聊天记录或配置文件。凭证需要在进程内短暂解密才能传给安装器，安装器随后存入系统凭证管理器。

### 启动新会话

安装并验证后退出原来的 Agent 会话，在终端执行：

```bash
erp-skill launch codex
# 或：
erp-skill launch claude-code
```

Codex 配置中的凭证变量名为 `ERP_OPEN_CODEX_KEY`；Claude Code 配置使用 `${ERP_OPEN_CLAUDE_KEY}`。`launch` 从系统凭证管理器读取 Key 并注入子进程，所以直接打开另一个普通客户端进程不一定能获得这些变量。若当前就在 Agent 中操作，应由用户退出后启动，不要让 Agent 嵌套启动一个交互会话。

安装器会安装共享 Skill 和 `references/`，并通过客户端自身命令注册 MCP。安装完成后请重启 Agent 或重连 MCP，使用当前会话的 `tools/list` 确认可见工具。

## 直接运行 codex 或 claude

MCP 的 `initialize` 是一次协议握手，本身不会安装 Skill，也不会保存客户端配置。下面说明以成功执行 `erp-skill install codex` 或 `erp-skill install claude-code` 为前提。

| 项目 | 新启动的 codex / claude 进程 | erp-skill launch |
| --- | --- | --- |
| 已安装的 ERP Skill | 可由客户端发现，相关请求触发后按需读取正文。 | 使用同一份已安装的 Skill。 |
| 已注册的 erp-open MCP 配置 | 客户端按自身设置和版本读取。 | 使用同一份持久化配置。 |
| 安装时保存在系统凭证管理器的 ERP Key | 不会调用本项目的凭证读取器。 | 读取 Key，只注入所启动的 Agent 子进程。 |
| 已鉴权的 ERP 工具调用 | 要求进程已获得对应环境变量，且凭证有效。 | 使用已保存凭证，仍需端点可达且服务端授权有效。 |

Codex 需要 `ERP_OPEN_CODEX_KEY`，Claude Code 需要 `ERP_OPEN_CLAUDE_KEY`。保存 Key 不会改变父终端的环境变量；运行一次 `launch` 也不会让其他终端或以后独立启动的客户端自动获得 Key。`doctor` 成功同样不会向终端导出凭证。

本项目当前将 Codex Skill 安装到 `$CODEX_HOME/skills/erp-open-platform`，未设置时为 `~/.codex/skills/erp-open-platform`；Claude Code 安装到 `~/.claude/skills/erp-open-platform`。Skill 的发现目录和禁用策略可能随客户端版本变化，未显示时应检查当前客户端的 Skill 列表，不能将此说明理解为所有未来版本的兼容承诺。

日常推荐 `erp-skill launch codex` / `erp-skill launch claude-code`。不要为了直接启动而把 Key 明文写进 shell 启动文件或 MCP 配置。本项目没有自动安装 shell alias，也不全局导出凭证。

参考：[Codex MCP](https://developers.openai.com/codex/mcp/)、[Codex Skills](https://developers.openai.com/codex/skills/)、[Claude Code MCP](https://code.claude.com/docs/en/mcp)、[Claude Code Skills](https://code.claude.com/docs/en/skills)。安装路径和系统凭证行为以上述项目实现为准。

## 连接 WorkBuddy

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

1. 在 WorkBuddy 的 Connector 导入入口选择生成的 `erp-open-platform-connector` 目录。
2. 在 Token 输入界面填写个人 API Key。
3. 连接或重连 MCP，并检查可用工具。

生成过程不需要提供 API Key，文件中只保留 `${ERP_OPEN_API_KEY}` 占位符。不要手工将真实 Key 替换进 JSON。WorkBuddy 不使用 `erp-skill launch`。

需要独立检查连通性时，可将前述安全输入示例中的安装命令改为：

```bash
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill doctor workbuddy --url "$MCP_URL" --key-stdin
```

检查完成后运行 `unset ERP_INSTALL_KEY`。PowerShell 同样可将安全输入示例中的 `install $Agent` 改为 `doctor workbuddy`。这一步不会配置 WorkBuddy 界面的 Token。

## 升级、回退与卸载

升级或回退：选择目标的已发布版本，使用对应脚本、安装包和校验清单重新安装 `erp-skill`，再检查 `erp-skill version`。

Codex / Claude Code 的 Skill 内容变化时，安装器会拒绝覆盖内容不同的已有文件。先备份自己修改过的 Skill；确认要替换后，按上述安全输入方式，在 `install` 命令末尾添加 `--force`。WorkBuddy 同理，确认后可为重新生成 Connector 的命令添加 `--force`，再重新导入。

Codex / Claude Code 可卸载当前端点的集成：

```bash
erp-skill uninstall codex --url https://erp-open.example.com/mcp
# 或：
erp-skill uninstall claude-code --url https://erp-open.example.com/mcp
```

WorkBuddy 请在其界面移除 Connector 和 Token。确认所有客户端都不再使用工具后，可以手动移除 `~/.local/bin/erp-skill` 或 Windows 对应的 `erp-skill.exe`。本地卸载不会吊销服务端 API Key；丢失设备或怀疑凭证泄露时，应联系管理员吊销或轮换。

## 常见问题

| 现象 | 处理方法 |
| --- | --- |
| `erp-skill` 命令不存在 | 将安装目录加入 PATH，或使用可执行文件的完整路径。 |
| 下载 404 | 核对版本、平台、文件名和镜像地址；只有正式上传的 Assets 才可下载。 |
| SHA-256 校验失败 | 停止安装，重新下载匹配版本的包和清单；不要绕过校验。 |
| Linux 找不到 `shasum` | 固定的 v0.1.1 脚本需要该命令；安装 Perl、改用手动校验，或使用仓库新版脚本。 |
| 凭证管理器不可用 | 确认 Keychain / 凭证管理器 / Secret Service 可用；不要改为明文保存 Key。 |
| 提示会覆盖已修改的 Skill / Connector | 先备份并检查已有文件，确认后才使用 `--force`。 |
| `INSTALLED_AUTHORIZATION_REQUIRED` | 初始化和工具发现成功，但缺少只读的 `get_platform_info` 授权，请联系管理员。 |
| 直接启动 codex / claude 能看到配置但调用失败 | 优先使用 erp-skill launch；直接启动不会读取本项目保存在系统凭证管理器的 Key。 |
| Agent 看不到新增工具 | 重连 MCP 或通过 `erp-skill launch` 启动新会话；以 `tools/list` 为准。 |
| WorkBuddy 只生成了目录 | 还需在客户端导入 Connector、填写 Token 并连接，不能据此认定安装已验证。 |

另见[安装提示词](agent-install-prompts.md)、[Agent 使用说明](agent-usage.md)及[错误处理](../references/errors.md)。
