# Agent usage

[中文](agent-usage.md) | English

Start with the [prebuilt installation guide](installation.en.md), or use the [agent installation prompts](agent-install-prompts.en.md). Codex, Claude Code, and WorkBuddy are supported integration targets.

## How the agent knows what ERP can do

1. MCP initialization instructions define identity, confirmation, and sensitive-data rules.
2. MCP `tools/list` returns the current Key's authorized tools, typed schemas, and business descriptions.
3. The shared Skill supplies search-then-select, read-before-write, idempotent creation, and confirmation workflows.

The current connection is authoritative. An agent must not infer an unavailable tool from documentation.

## Describe business goals in either language

- “搜索姓名包含张学的员工，并列出工号和部门。” / “Find employees named Zhang Xue and show their employee numbers and departments.”
- “先查询订单当前状态，再告诉我是否可以提交。” / “Read the order's current state before explaining whether submission is possible.”
- “创建销售单前先展示客户、商品、数量和金额，等我确认后再执行。” / “Show the customer, products, quantities, and amounts before creating a sales document; wait for my confirmation.”

The agent should match your language. When searches return multiple candidates, show distinguishing fields and let the user choose; when there are none, report that fact. Explain the exact target and business impact before high-impact financial, approval, posting, or shipment operations.

## Authorization and identity

After an administrator grants capabilities to the same Key, reconnect MCP or start a new agent session and inspect `tools/list`. There is no local sync command, generated tool inventory, or background updater.

The server checks revocations on every call. A stale client list cannot bypass authorization. The ERP user and tenant are selected when the Key is issued; tool arguments cannot override them. Downstream services still enforce business permissions and data visibility.

## Initialization, installation, and startup

MCP `initialize` is a protocol handshake. It does not by itself install the shared Skill or persist client configuration.

`erp-skill install` installs the Skill and persists MCP registration. Starting the same locally configured client under the same user makes that configuration discoverable. The Skill's full instructions are activated when relevant, rather than necessarily applied to every request.

Bare `codex` or `claude` does not run this project's OS credential-manager loader. Unless the new process already has the matching `ERP_OPEN_CODEX_KEY` / `ERP_OPEN_CLAUDE_KEY`, discovering the MCP configuration is insufficient for authenticated ERP access. Installing or running `doctor` does not export these variables to the terminal.

Use `erp-skill launch codex` or `erp-skill launch claude-code` in a terminal for routine startup. It reads the saved Key and injects it only into the child process. A separate terminal or later bare client process does not inherit that injection. See the [startup FAQ](installation.en.md#starting-codex-or-claude-directly).

WorkBuddy requires importing the generated Connector, entering the Key in its Token UI, and connecting MCP. It does not use `launch`; generating a directory alone is not a completed connection.

## Verification results

`doctor` performs `initialize`, reads every page of `tools/list`, and invokes only read-only `get_platform_info` when authorized. It does not invoke write-capable business tools.

- `VERIFIED`: initialization, discovery, and platform-information verification succeeded.
- `INSTALLED_AUTHORIZATION_REQUIRED`: initialization and discovery succeeded, but `get_platform_info` is not exposed; ask an administrator for authorization.

Do not create a business document merely to test installation.

## Recorded validation scope

These are historical test records, not a claim that every latest client version has been tested:

- As of 2026-10-30, Codex CLI `0.163.0` MCP add/remove syntax and local adapter tests passed on macOS.
- Claude Code `2.1.304` had corresponding command and adapter checks at that time.
- WorkBuddy's Connector uses Token mode and declares minimum version `4.23.0`; WorkBuddy was not installed in that implementation environment, so UI import was not verified there.
- Releases are cross-compiled for macOS arm64/amd64, Linux arm64/amd64, and Windows amd64. Cross-compilation does not prove native Windows, Linux, or WorkBuddy installation.

See [troubleshooting](installation.en.md#troubleshooting) and [error handling](../references/errors.en.md).
