# 让 Agent 帮你安装：可复制的中文提示词

将对应段落复制到 Codex、Claude Code 或 WorkBuddy 中。把 MCP 地址替换为管理员提供的真实地址；**不要把个人 API Key 放在提示词中**。

这些提示词将“下载预编译工具”和“配置 Agent”分成两个实际步骤。能访问终端、网络和本机文件的 Agent 可以自动完成下载、校验和生成配置；凭证输入、客户端导入或重启应由用户在安全界面完成。没有终端权限时，应提供本机执行命令，不能宣称已经安装。

## Codex 安装提示词

```text
请为当前用户安装 ERP Agent Skill 并配置当前电脑上的 Codex。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.1
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查操作系统、处理器架构、网络和 codex 命令。优先使用 Release 预编译包，不要默认安装 Go、克隆或编译源码。
2. 从 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1 选择对应平台包，下载 checksums.txt，核对 SHA-256 后再安装到当前用户可用的目录。
   也可使用该 Tag 下 scripts/install.sh 或 scripts/install.ps1；v0.1.1 脚本需指定下载目录 https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1。根据操作系统检查脚本所需工具，校验失败立即停止。
3. 确保 PATH 可找到 erp-skill，运行 erp-skill version，核对版本为 0.1.1。
4. 不要向我索取聊天中的 API Key，不要把 Key 写进命令参数或文件。需要凭证时，给我一段在本机终端隐藏输入并通过标准输入执行以下命令的代码：
   erp-skill install codex --url <真实MCP地址> --key-stdin
5. 安装完成后运行 erp-skill doctor codex --url <真实MCP地址>。只验证连接和只读能力，不执行写入业务操作。
6. 不要覆盖我已有的修改；如果安装器提示冲突，说明文件和影响，先让我备份并确认是否使用 --force。
7. 告诉我退出当前 Codex 会话，在终端用 erp-skill launch codex 启动新会话，以加载 Skill、MCP 和系统凭证。
8. 用中文报告安装路径、版本、doctor 实际结果和剩余步骤。尚未输入凭证或尚未启动新会话时，不能把状态说成已验证。
```

安装后可发送：“列出当前 ERP MCP 连接可用的只读能力，并解释我的用户和数据范围；不要执行写入。”

## Claude Code 安装提示词

```text
请为当前用户安装 ERP Agent Skill 并配置当前电脑上的 Claude Code。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.1
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查操作系统、处理器架构、网络和 claude 命令。优先下载 Release 预编译包，不要默认安装 Go 或从源码编译。
2. 在 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1 下载匹配的平台包和 checksums.txt，核对 SHA-256 后安装 erp-skill。
   也可运行对应 Tag 的 scripts/install.sh 或 scripts/install.ps1，并指定下载目录 https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1。校验失败立即停止。
3. 设置当前用户 PATH，运行 erp-skill version，核对为 0.1.1。
4. 不在聊天里收集个人 API Key。给我本机终端隐藏输入的代码，将 Key 通过标准输入交给：
   erp-skill install claude-code --url <真实MCP地址> --key-stdin
   不要将真实 Key 写进配置、命令参数或仓库。Claude 配置中的 ${ERP_OPEN_CLAUDE_KEY} 必须保留为占位符。
5. 安装后运行 erp-skill doctor claude-code --url <真实MCP地址>，只做连接与只读验证。
6. 如已有修改导致冲突，先说明并让我备份、确认，再考虑 --force。
7. 告诉我退出当前 Claude Code，在终端运行 erp-skill launch claude-code 启动新会话；不要嵌套启动交互进程。
8. 用中文报告安装路径、版本、验证结果和未完成事项，不要把待输入凭证或待重启的状态说成已经全部完成。
```

## WorkBuddy 安装提示词

```text
请为当前用户准备 ERP Agent Skill 的 WorkBuddy Connector，并指导我完成导入。

官方仓库：https://github.com/xsy-tech/erp-agent-skills
目标发布版本：v0.1.1
MCP 地址：https://erp-open.example.com/mcp（执行前请让我替换成真实地址）

请执行：
1. 检查操作系统、处理器架构，以及是否有终端和网络权限。优先使用 Release 预编译包，不要默认从源码编译。
2. 从 https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1 下载匹配的安装包和 checksums.txt，核对 SHA-256 后安装 erp-skill，设置 PATH 并核对版本为 0.1.1。
3. 运行：
   erp-skill install workbuddy --url <真实MCP地址> --output ./erp-open-platform-connector
   告诉我生成目录的绝对路径；如果目录已有修改，先说明并让我备份确认，不要直接 --force。
4. 生成 Connector 不需要个人 API Key。不要在聊天里询问 Key，也不要把真实 Key 写进 JSON；保留 ${ERP_OPEN_API_KEY} 占位符。
5. 指导我在 WorkBuddy 的 Connector 导入入口导入生成目录，再在 Token 界面输入个人 API Key，然后连接或重连 MCP。Release 中的 ZIP / tar.gz 是命令行工具包，不能代替生成的 Connector 目录直接导入。
6. 如果当前客户端不能自动完成导入或 Token 输入，请明确列出我需要操作的步骤。不要尝试 erp-skill launch workbuddy，这个命令不受支持。
7. 如果我需要独立的命令行连通性检查，给出本机隐藏输入 Key 并通过标准输入执行 erp-skill doctor workbuddy --url <真实MCP地址> --key-stdin 的代码；说明这不会替我设置 WorkBuddy 的 Token。
8. 用中文报告二进制版本、Connector 路径、实际检查结果及导入/连接状态；只生成目录不能视作已验证成功。
```

## 如何判断完成

- `erp-skill version` 与选定发布版本一致。
- Codex / Claude Code：安装器成功，`doctor` 返回真实检查结果，并通过 `launch` 启动新会话。
- WorkBuddy：已生成 Connector，完成界面导入、Token 配置及 MCP 连接。独立 `doctor` 检查只能证明其使用的地址和 Key 的连通性。
- `INSTALLED_AUTHORIZATION_REQUIRED` 需要管理员补充授权，不能等同于完整的 `VERIFIED`。

手动安装和安全输入的完整命令见[安装指南](installation.md)。
