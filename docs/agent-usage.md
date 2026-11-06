# Agent 使用说明

首次使用先按[安装指南](installation.md)安装预编译包并配置客户端，也可以复制[安装提示词](agent-install-prompts.md)让 Agent 协助。支持 Codex、Claude Code 和 WorkBuddy。

## Agent 如何知道可以做什么

ERP 能力信息有三层：

1. MCP 初始化指令规定全局身份、确认和敏感信息规则。
2. MCP `tools/list` 返回当前个人 API Key 获授权的工具、类型定义和业务描述。
3. 共享 Skill 提供先搜索再选定、先读再写、幂等创建、提交前确认等可复用工作流。

Agent 不能根据文档猜测一个当前连接中不存在的工具。实际可调用的能力以 MCP 会话为准。

## 用中文描述业务目标

不需要猜测工具名，直接描述需要完成的事情：

- “搜索姓名包含张学的员工，并列出工号和部门。”
- “先查询订单当前状态，再告诉我是否可以提交。”
- “创建销售单前先展示客户、商品、数量和金额，等我确认后再执行。”

搜索结果有多个候选项时，Agent 应展示区分字段并请你选定；没有结果时应如实说明，不应编造数据。财务、审批、过账、发货等高影响操作，应先说明对象和影响，再按安全规则确认。

## 权限变化与个人身份

管理员给同一个 Key 增加能力后，重连 MCP 或启动新 Agent 会话，即可在 `tools/list` 看到新工具。没有本地 `sync` 命令、生成的本地工具清单或后台更新进程。

撤权由服务端逐次调用校验。旧的客户端工具列表不能绕过服务端授权。

Key 绑定的 ERP 用户和租户在开放平台管理端确定，工具参数不能切换该身份。下游 ERP 服务仍会校验业务权限和数据范围。

## 连接与启动

Codex 和 Claude Code 完成安装后，应从终端运行 `erp-skill launch codex` 或 `erp-skill launch claude-code` 启动新会话。启动器从系统凭证管理器取出 Key，只注入目标子进程。

WorkBuddy 需要导入生成的 Connector，在其 Token 界面填写个人 Key 并连接 MCP。它不使用 `erp-skill launch`，仅生成 Connector 目录也不代表已经连接成功。

## 验证结果说明

`erp-skill doctor` 执行 MCP `initialize`，读取所有分页的 `tools/list`，并在已授权时调用只读的 `get_platform_info`，不会执行具有写入能力的业务工具。

- `VERIFIED`：初始化、工具发现和平台信息检查成功。
- `INSTALLED_AUTHORIZATION_REQUIRED`：初始化和工具发现成功，但当前 Key 未暴露 `get_platform_info`，需要管理员补充授权。

不要用“尝试创建一张单据”作为安装验证步骤。

## 已有验证范围

以下是项目此前记录的验证信息，不表示最新客户端版本都已完成测试：

- 截至 2026-10-30，Codex CLI `0.163.0` 的 MCP 注册/移除命令和本地适配器测试在 macOS 上完成。
- 同期 Claude Code `2.1.304` 完成对应命令与适配器测试。
- WorkBuddy Connector 使用 Token 模式，声明最低版本 `4.23.0`；当时实现环境没有安装 WorkBuddy，尚未完成客户端实际导入测试。
- Release 支持 macOS arm64/amd64、Linux arm64/amd64、Windows amd64 的交叉编译。交叉编译成功不等于在 Windows、Linux 或 WorkBuddy 中完成原生安装验证。

问题排查见[安装指南](installation.md#常见问题)和[错误处理](../references/errors.md)。
