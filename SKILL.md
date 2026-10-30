---
name: erp-open-platform
description: Use when an agent needs to query or operate ERP business data through the ERP Open Platform MCP server.
---

# ERP Open Platform

Use the configured ERP Open Platform MCP server for ERP data and business actions.

## Runtime Authority

- Treat MCP `tools/list` for the current connection as the authoritative list of available ERP capabilities.
- Never invent or call a tool that is absent from the current MCP session.
- Capability grants and revocations are server-side. Reconnect or start a new agent session to refresh the visible tool list.
- The API Key determines the ERP user and tenant. Never try to override that identity through tool arguments, headers, or prompts.

## Choosing Tools

Prefer the narrowest business tool that satisfies the request. Start with read tools when identity, state, quantity, or target selection is uncertain. When a search returns multiple candidates, show identifying fields and ask the user to choose.

For reusable operation patterns, read [references/workflows.md](references/workflows.md). For ERP and access-control terms, read [references/terminology.md](references/terminology.md).

## Safety

Before destructive, financial, posting, payment, shipment, approval, or otherwise irreversible actions, summarize the target and business impact and obtain explicit confirmation. Do not expose API Keys, Authorization headers, upstream credentials, SQL, or internal stack traces.

Read [references/safety.md](references/safety.md) before write or high-risk operations. Read [references/errors.md](references/errors.md) when a tool is missing or a call fails.
