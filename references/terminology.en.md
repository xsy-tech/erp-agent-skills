# ERP terminology

[中文](terminology.md) | English

| Term | Meaning |
| --- | --- |
| API Key | Personal data-plane credential bound by the server to one ERP user, tenant, Client, and capability set. |
| Client | Application or person-facing integration subject that owns quotas and API Keys. |
| Capability | Server-side permission to invoke one specific OpenAPI operation and its matching MCP tool. |
| Tenant | ERP data-isolation boundary chosen when the Key is issued. |
| Bound user | ERP employee identity used by downstream services for business permissions and data-scope checks. |
| Tool | Typed MCP operation currently visible in tools/list for the connected Key. |
| Idempotency key | Caller-provided operation identifier that prevents duplicate effects for supported writes. |

Capability authorization permits an operation to be attempted. Downstream ERP services still enforce the bound user's business permissions and data visibility.
