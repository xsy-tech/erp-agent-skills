# v0.1.2 project1 验证记录

中文 | [English](verification-v0.1.2-project1.en.md)

验证日期：2026-11-06。**整体部分通过**：Codex 与 WorkBuddy 已通过原生模型会话查询真实 ERP 数据；Claude Code 的模型请求被 403 阻塞，另有一项上游服务健康检查失败。

## 验证范围

- 使用 [v0.1.2 官方 Release](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2) 的 macOS arm64 安装包，没有重新编译测试二进制。
- 安装包 SHA-256：`1a7c9e62e2b618d824757eb57e08c014c1a83ce5336a9d08f4def4ea8041ea95`。
- 客户端：Codex CLI 0.166.0、Claude Code 2.1.316、WorkBuddy 桌面版 5.3.11。
- 本机开放平台使用 project1 配置，鉴权后的 `get_platform_info` 返回 `environment=project1`。本轮使用真实服务，没有业务 Mock。
- 同一长期测试 Key 已持久化到三个客户端的用户级 MCP 配置，权限限于平台信息、服务状态和用户搜索三项只读能力。到期日为 2027-11-06；最初的短期测试 Key 已撤销。
- 未测试其他操作系统、CPU 架构，以及 Codex/Claude 的桌面界面。

## 结果

| 检查项 | Codex | Claude Code | WorkBuddy 桌面版 |
| --- | --- | --- | --- |
| 使用已保存 Key 执行 doctor | VERIFIED | VERIFIED | VERIFIED |
| 原生 MCP 连接与 Skill 发现 | 通过 | 通过 | 完成首次信任后通过 |
| 模型调用平台信息，确认 project1 | 通过 | 模型 403，未执行 | 通过 |
| 模型调用用户搜索 | 通过 | 模型 403，未执行 | 通过 |
| api-user 健康检查 | 独立 MCP 对照通过 | 未执行 | 通过 |
| saas-erp-backend 健康检查 | 失败，上游不可用 | 未执行 | 失败，上游不可用 |

Codex 的实际模型会话调用了平台信息、用户搜索和服务状态工具。WorkBuddy 的实际桌面会话调用了平台信息、用户搜索以及两个目标的服务状态工具；四条服务端审计对应三次成功和一次上游失败。本轮没有调用业务写入工具。

Claude Code 原生初始化已确认 ERP MCP 连接成功、Skill 已发现，但模型请求返回 `403 Request not allowed`，无法继续验证模型发起业务调用。该阻塞与 ERP Key 鉴权分开记录。

`saas-erp-backend` 的配置目标在验证时没有监听，因此健康检查记为失败。没有修改业务代码或模拟健康响应来获得通过结果。

## WorkBuddy 首次连接

直接安装 MCP 和 Skill 后，WorkBuddy 可能仍要求信任该 MCP 服务。本轮第一轮会话发现了 Skill，但无法发现 ERP 工具；进入 **MCP 服务管理** 后完成首次信任，界面显示 **3/3 个工具已启用**，再用新会话复测成功。

无需导入 Connector 包或使用 `erp-skill launch`。`doctor` 验证的是保存的 MCP 配置、鉴权与只读调用；客户端自己的信任状态和模型账号是否可用，还需在原生客户端验证。

## 记录与配置保留

长期 Key、用户级 MCP 配置、Skill 和 WorkBuddy 信任状态保留供后续测试；配置文件权限已核对为 `0600`，原有其他 MCP 和设置已保留。临时凭据副本已清理，本轮管理会话已退出。

仓库仅发布此脱敏摘要。原始日志、截图、服务端审计及身份元数据保留在本地，不包含在提交中。
