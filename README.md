# ERP Agent Skill Bundle

`erp-agent-skill` distributes one shared ERP Skill and the adapters required to connect Codex, Claude Code, and WorkBuddy to ERP Open Platform MCP.

## Operating Model

- MCP `tools/list` is the live authority for the tools available to the current personal API Key.
- The Skill provides stable ERP terminology, workflows, safety rules, and error handling. It does not contain a generated tool catalog.
- Each API Key is bound server-side to one Client, ERP user, tenant, and capability set.
- Codex and Claude Code keys are stored in the operating-system credential manager and injected only into a launched child process.
- WorkBuddy stores the personal key through its Token UI; generated connector files contain only `${ERP_OPEN_API_KEY}`.
- New grants appear after MCP reconnect or a new agent session. No local capability sync command or daemon is required.

## Build

```bash
make check
make build
make dist
```

## Codex

```bash
printf '%s\n' "$ERP_OPEN_API_KEY" | \
  erp-skill install codex \
  --url https://erp-open.example.com/mcp \
  --key-stdin

erp-skill doctor codex --url https://erp-open.example.com/mcp
erp-skill launch codex
```

The installer registers `erp-open` with `ERP_OPEN_CODEX_KEY` as the bearer-token environment variable. The variable value is supplied only by `erp-skill launch codex`.

## Claude Code

```bash
printf '%s\n' "$ERP_OPEN_API_KEY" | \
  erp-skill install claude-code \
  --url https://erp-open.example.com/mcp \
  --key-stdin

erp-skill doctor claude-code --url https://erp-open.example.com/mcp
erp-skill launch claude-code
```

The Claude MCP configuration stores the literal placeholder `${ERP_OPEN_CLAUDE_KEY}`, not the credential value.

## WorkBuddy

```bash
erp-skill install workbuddy \
  --url https://erp-open.example.com/mcp \
  --output ./erp-open-platform-connector
```

Import the generated connector and enter the personal API Key in WorkBuddy's Token UI. For pre-distribution connectivity validation:

```bash
printf '%s\n' "$ERP_OPEN_API_KEY" | \
  erp-skill doctor workbuddy \
  --url https://erp-open.example.com/mcp \
  --key-stdin
```

## Credential Rules

- Never pass a key through `--key`; that option is intentionally unsupported.
- Never commit keys, profiles, generated shell history, or copied Authorization headers.
- Revoke or rotate a key in ERP Open Platform when a device or account is lost.
- The local endpoint profile contains only the MCP URL and is not a capability cache.

See [agent usage](docs/agent-usage.md) and [enterprise release](docs/enterprise-release.md) for operational details.
