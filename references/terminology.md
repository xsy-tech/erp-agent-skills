# ERP Terminology

- **API Key**: Personal data-plane credential bound by the server to one ERP user, tenant, Client, and capability set.
- **Client**: Application or person-facing integration subject that owns quotas and API Keys.
- **Capability**: Server-side permission to invoke one specific OpenAPI operation and its matching MCP Tool.
- **Tenant**: ERP data-isolation boundary selected when the API Key is issued.
- **Bound user**: ERP employee identity used by downstream services for business permission and data-scope checks.
- **Tool**: Typed MCP operation currently visible through `tools/list` for the connected API Key.
- **Idempotency key**: Caller-provided identifier that prevents duplicate effects for supported write operations.

Capability permission only allows the operation to be attempted. Downstream ERP services still enforce the bound user's business permissions and data visibility.
