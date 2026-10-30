# Agent Usage

## How The Agent Knows What ERP Can Do

ERP capability understanding has three layers:

1. MCP initialize instructions define global identity, confirmation, and sensitive-data rules.
2. MCP `tools/list` returns the current API Key's authorized tools with typed schemas and business descriptions.
3. The shared Skill supplies reusable ERP workflows such as search-then-select, read-before-write, idempotent create, and confirm-before-submit.

The Agent must never infer a missing Tool from this document. The current MCP connection is authoritative.

## Capability Changes

When an administrator grants another Capability to the same Key, reconnect the MCP server or start a new Agent session. The new Tool is then returned by `tools/list`. There is no `sync`, generated local Tool file, polling daemon, or background updater.

Revocations are enforced by the server on every Tool call. An old client-side list cannot bypass authorization.

## Personal Identity

The Key's bound ERP user and tenant are selected in the ERP Open Platform management console. Tool arguments cannot switch this identity. Downstream ERP services continue to enforce that user's business permissions and data visibility.

## Safe Requests

Good requests identify the business goal rather than guessing a Tool name:

- “搜索姓名包含张学的员工，并列出工号和部门。”
- “先查询订单当前状态，再告诉我是否可以提交。”
- “创建销售单前先展示客户、商品、数量和金额，等我确认后再执行。”

For multiple search matches, the Agent should show distinguishing fields and ask the user to select. For no matches, it must report no data instead of inventing a result.

## Tested Tooling

As of October 30, 2026:

- Codex CLI `0.163.0`: MCP add/remove command syntax and local adapter tests completed on macOS.
- Claude Code `2.1.304`: MCP add/remove command syntax and local adapter tests completed on macOS.
- WorkBuddy: Connector structure follows Token mode with `min_workbuddy_version` `4.23.0`; WorkBuddy runtime was not installed in the implementation environment.

Release builds are cross-compiled for macOS arm64/amd64, Linux arm64/amd64, and Windows amd64. Native installation testing is still required on representative Windows and Linux hosts before enterprise-wide rollout.
