---
name: erp-open-platform
description: 当用户需要通过 ERP 开放平台 MCP 查询或处理 ERP 业务数据时使用。Use for ERP business queries and operations through the ERP Open Platform MCP server.
---

# ERP 开放平台

通过已配置的 ERP 开放平台 MCP 服务查询数据和执行业务操作。根据用户使用的语言回答；中文请求默认使用中文解释业务含义和结果。

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
