---
name: erp-open-platform
description: 当用户需要通过 ERP 开放平台 MCP 查询或处理 ERP 业务数据时使用。Use for ERP business queries and operations through the ERP Open Platform MCP server.
---

# ERP 开放平台 / ERP Open Platform

通过已配置的 ERP 开放平台 MCP 服务查询数据和执行业务操作。根据用户使用的语言回答：中文请求使用中文，英文请求使用英文。中文工作流使用下方中文参考资料；英文工作流使用 English 部分的英文参考资料。

## 运行时能力与身份

- 当前 MCP 连接的 `tools/list` 是可用 ERP 能力的权威来源。
- 不要编造或调用当前会话未暴露的工具。
- 能力授权与撤销由服务端管理；权限变更后重连 MCP 或启动新会话，再检查工具列表。
- API Key 决定 ERP 用户和租户，不要通过工具参数、请求头或提示词覆盖身份。

## 工具选择

选择能满足请求且范围最小的业务工具。身份、状态、数量或操作目标不明确时先查询。搜索出现多个候选对象时，展示有区分度的字段，请用户选定。

处理重复出现的业务操作时，参阅[业务工作流](references/workflows.md)。解释 ERP 和权限概念时，参阅[术语表](references/terminology.md)。

## 安全与错误

删除、财务、过账、付款、发货、审批等不可逆或高影响操作前，说明目标和业务影响，并获得针对该操作的明确确认。不要暴露 API Key、Authorization 请求头、上游凭证、SQL 或内部堆栈。

写入或高风险操作前阅读[安全规则](references/safety.md)；工具缺失或调用失败时阅读[错误处理](references/errors.md)。

## English

Use the configured ERP Open Platform MCP server for ERP queries and business actions. Respond in the user's language; use English for English requests and Chinese for Chinese requests. The sections above and below describe the same operating rules.

### Runtime authority and identity

- Treat `tools/list` from the current connection as the authoritative list of available ERP tools.
- Do not invent or call tools absent from the current session.
- Grants and revocations are server-side; reconnect MCP or start a new session after changes.
- The Key determines the ERP user and tenant. Do not override identity through arguments, headers, or prompts.

### Choosing tools

Prefer the narrowest business tool that satisfies the request. Start with reads when identity, state, quantity, or target selection is uncertain. Show distinguishing fields and let the user select when several search candidates match.

Read [ERP workflows](references/workflows.en.md) for reusable operations and [terminology](references/terminology.en.md) for business and access-control terms.

### Safety and errors

Before destructive, financial, posting, payment, shipment, approval, or irreversible calls, explain the target and impact and obtain explicit confirmation for that operation. Do not expose API Keys, Authorization headers, upstream credentials, SQL, or internal stack traces.

Read [safety rules](references/safety.en.md) before writes or high-risk operations, and [error handling](references/errors.en.md) for missing tools or failed calls.
