# 让 Agent 帮你安装：可复制的中文提示词

中文 | [English](agent-install-prompts.en.md)

将对应段落复制到 Codex、Claude Code 或 WorkBuddy 中。把 MCP 地址替换为管理员提供的真实地址；**不要把个人 API Key 放在提示词中**。

这些提示词将“下载预编译工具”和“配置 Agent”分成两个实际步骤。能访问终端、网络和本机文件的 Agent 可以自动完成下载、校验和生成配置；凭证输入、客户端导入或重启应由用户在安全界面完成。没有终端权限时，应提供本机执行命令，不能宣称已经安装。

## Codex 安装提示词

```text
请为当前用户安装 ERP Agent Skill 并配置当前电脑上的 Codex。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.2
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查操作系统、处理器架构、网络和 codex 命令。优先使用 Release 预编译包，不要默认安装 Go、克隆或编译源码。
2. 从 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2 选择对应平台包，下载 checksums.txt，核对 SHA-256 后再安装到当前用户可用的目录。
   也可使用该 Tag 下 scripts/install.sh 或 scripts/install.ps1；可指定下载目录 https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.2。根据操作系统检查脚本所需工具，校验失败立即停止。
3. 确保 PATH 可找到 erp-skill，运行 erp-skill version，核对版本为 0.1.2。
4. 不要向我索取聊天中的 API Key，不要把 Key 写进命令参数或共享文件；新版安装器会将 Key 保存到本机用户级私有 MCP 配置。需要凭证时，给我一段在本机终端隐藏输入并通过标准输入执行以下命令的代码：
   erp-skill install codex --url <真实MCP地址> --key-stdin
5. 安装完成后运行 erp-skill doctor codex --url <真实MCP地址>。只验证连接和只读能力，不执行写入业务操作。
6. 不要覆盖我已有的修改；如果安装器提示冲突，说明文件和影响，先让我备份并确认是否使用 --force。
7. 先核对安装器版本：包含持久化鉴权支持的新版安装器，安装后重启客户端，再直接运行 codex 或打开对应本地桌面 Code 会话。旧 v0.1.1 仍使用 erp-skill launch codex。不要把旧版宣称为新版行为，也不要嵌套启动交互会话。
8. 用中文报告安装路径、版本、doctor 实际结果和剩余步骤。尚未输入凭证或尚未启动新会话时，不能把状态说成已验证。
```

v0.1.2 会持久化 Skill 和包含鉴权的用户级 MCP 配置，之后可正常启动本地客户端。旧 v0.1.1 用户升级后需重新安装集成，才能迁移原有启动流程。

安装后可发送：“列出当前 ERP MCP 连接可用的只读能力，并解释我的用户和数据范围；不要执行写入。”

## Claude Code 安装提示词

```text
请为当前用户安装 ERP Agent Skill 并配置当前电脑上的 Claude Code。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.2
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查操作系统、处理器架构、网络和 claude 命令。优先下载 Release 预编译包，不要默认安装 Go 或从源码编译。
2. 在 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2 下载匹配的平台包和 checksums.txt，核对 SHA-256 后安装 erp-skill。
   也可运行对应 Tag 的 scripts/install.sh 或 scripts/install.ps1，并指定下载目录 https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.2。校验失败立即停止。
3. 设置当前用户 PATH，运行 erp-skill version，核对为 0.1.2。
4. 不在聊天里收集个人 API Key。给我本机终端隐藏输入的代码，将 Key 通过标准输入交给：
   erp-skill install claude-code --url <真实MCP地址> --key-stdin
   不要把真实 Key 写进聊天、进程参数、项目配置或仓库。新版安装器会将 Authorization 写入本机用户级私有配置；旧 v0.1.1 仍使用 ${ERP_OPEN_CLAUDE_KEY} 占位符，不要混淆两个版本的行为。
5. 安装后运行 erp-skill doctor claude-code --url <真实MCP地址>，只做连接与只读验证。
6. 如已有修改导致冲突，先说明并让我备份、确认，再考虑 --force。
7. 使用包含持久化鉴权的新版本时，安装后重启客户端，直接运行 claude 或打开 Claude Code 本地桌面 Code 会话即可。若使用旧 v0.1.1，则仍执行 erp-skill launch claude-code。不要嵌套启动交互进程。
8. 用中文报告安装路径、版本、验证结果和未完成事项，不要把待输入凭证或待重启的状态说成已经全部完成。
```

## WorkBuddy 安装提示词

以下直接安装流程使用 v0.1.2 预编译包，不能使用仅支持 Connector 导出的旧 v0.1.1 代替。

```text
请为当前用户直接安装 ERP MCP 和 ERP Skill 到 WorkBuddy，不要求导入 Connector 包。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.2。不要静默回退到仅导出 Connector 的 v0.1.1。
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查系统、处理器、终端和网络权限，以及 WorkBuddy 是否已安装。从 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2 下载匹配平台的预编译包和 checksums.txt，核对 SHA-256 和 erp-skill version 为 0.1.2；下载或校验失败应停止，不要默认安装 Go 或从源码编译。
2. 不在聊天中询问 API Key。给我本机终端隐藏输入代码，通过标准输入运行：
   erp-skill install workbuddy --url <真实MCP地址> --key-stdin
   此模式不要加 --output，也不需要我填写 Connector Token 表单。
3. 安装器会更新 ~/.workbuddy/mcp.json 的 erp-open，使用 type=streamableHttp 和 headers.Authorization 保存鉴权，并安装 ~/.workbuddy/skills/erp-open-platform。配置是含个人 Key 的用户级私有文件，不得写到共享项目配置、仓库或聊天，不得展示完整内容。
4. 如遇 Skill 冲突，先说明并让我备份确认，再考虑 --force。不要删除其他 MCP 或修改 WorkBuddy 的其他设置。
5. 运行 erp-skill doctor workbuddy --url <真实MCP地址>，读取保存的 Key 做只读验证。无需再次 --key-stdin，除非我明确需要独立检查另一把 Key。
6. 如已启用旧的 ERP Connector，先指导我停用，避免与新 erp-open 重复暴露工具；不要自动删除旧 Connector。
7. 告诉我重启 WorkBuddy 或在自定义 MCP 页面重连 erp-open，确认 ERP Skill 启用。不要使用 erp-skill launch workbuddy。
8. 用中文报告版本、安装路径、实际 doctor 状态及待完成的重连步骤。文件写入和协议检查成功不能当作 WorkBuddy 界面实际验收成功。
```

如确实需要 Connector 包分发，可使用 `erp-skill install workbuddy --url <MCP地址> --output ./erp-open-platform-connector`。此模式不接收 Key、不配置本机 MCP，也不安装本机 Skill；导入后由 Token 表单处理凭证。它与默认直接安装二选一。

## 如何判断完成

- `erp-skill version` 与选定发布版本一致。
- Codex / Claude Code：安装器成功，`doctor` 返回真实检查结果，并按所用版本重启客户端；新版可直接使用，旧 v0.1.1 使用 launch。
- WorkBuddy 默认直接安装：用户级 MCP 和 Skill 已写入，doctor 使用保存的凭证完成只读验证，客户端重连后确认 MCP 和 Skill 可用。可选 Connector 导出模式另需界面导入和 Token 配置。
- `INSTALLED_AUTHORIZATION_REQUIRED` 需要管理员补充授权，不能等同于完整的 `VERIFIED`。

手动安装和安全输入的完整命令见[安装指南](installation.md)。
