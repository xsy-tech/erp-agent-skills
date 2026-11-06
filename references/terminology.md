# ERP 术语表

| 术语 | 含义 |
| --- | --- |
| API Key（个人密钥） | 服务端签发的个人数据访问凭证，绑定一个 ERP 用户、租户、Client 及能力集合。 |
| Client（接入主体） | 持有配额和 API Key 的应用或面向人员的集成主体。 |
| Capability（能力权限） | 调用某个具体 OpenAPI 操作及对应 MCP Tool 的服务端授权。 |
| Tenant（租户） | ERP 的数据隔离边界，在签发 API Key 时确定。 |
| Bound user（绑定用户） | 下游 ERP 服务进行业务权限与数据范围校验时使用的员工身份。 |
| Tool（工具） | 当前 Key 在 MCP `tools/list` 中可见、带有类型定义的业务操作。 |
| Idempotency key（幂等键） | 调用方提供的操作标识，用于防止支持幂等的写操作重复产生业务效果。 |

Capability 授权只表示允许尝试某项操作。下游 ERP 服务仍会校验绑定用户的业务权限和可见数据范围。
