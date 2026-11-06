# Installation prompts for your agent

[中文](agent-install-prompts.md) | English

Paste the relevant prompt into Codex, Claude Code, or WorkBuddy. Replace the MCP URL before execution. Never paste a personal API Key into the prompt.

An agent needs terminal, network, and filesystem access to perform these steps. Without that access, it should provide local commands and report the remaining actions. Hidden credential entry, client UI import, and restarting an interactive session may require the user.

## Codex

```text
Install ERP Agent Skill for the current user and configure Codex on this computer.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.1
MCP URL: https://erp-open.example.com/mcp (ask me to replace this placeholder before execution)

1. Check OS, processor architecture, network access, and the codex executable. Prefer the matching prebuilt Release archive. Do not install Go or compile source by default.
2. Download the matching archive and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1. Verify SHA-256 before installation. Alternatively, use that Tag's scripts/install.sh or scripts/install.ps1 with the base URL https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1, after checking its dependencies. Stop on checksum failure.
3. Add the current user's installation directory to PATH and verify erp-skill version reports 0.1.1.
4. Do not request a Key in chat or place it in files or literal command arguments. Give me local hidden-input code that passes it through stdin to:
   erp-skill install codex --url <real-MCP-URL> --key-stdin
5. Run erp-skill doctor codex --url <real-MCP-URL> after installation, using only connection and read-only verification.
6. If managed files conflict, explain the affected files and ask me to back them up and confirm replacement before using --force.
7. Tell me to exit this Codex session and run erp-skill launch codex in a terminal. Bare codex can discover installed configuration but does not automatically load the ERP Key from this project's credential manager. Do not start a nested interactive session.
8. Report the installation path, version, actual verification result, and remaining user steps. Missing credential input or a required restart must not be reported as verified success.
```

After setup, ask: "List the read-only capabilities in the current ERP MCP connection and explain my identity and data scope. Do not write business data."

## Claude Code

```text
Install ERP Agent Skill for the current user and configure Claude Code on this computer.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.1
MCP URL: https://erp-open.example.com/mcp (ask me to replace this placeholder before execution)

1. Check OS, architecture, network access, and the claude executable. Prefer a prebuilt Release archive over a source build.
2. Download the matching package and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1. Verify SHA-256. The corresponding Tag's installer script can also be used with base URL https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1. Stop if verification fails.
3. Set the user's PATH and confirm erp-skill version reports 0.1.1.
4. Do not collect a Key in chat. Provide local hidden-input code that passes the Key through stdin to:
   erp-skill install claude-code --url <real-MCP-URL> --key-stdin
   Keep ${ERP_OPEN_CLAUDE_KEY} as a placeholder in the Claude configuration. Do not put a literal Key into files, command arguments, or the repository.
5. Run erp-skill doctor claude-code --url <real-MCP-URL> using read-only checks.
6. Explain existing-file conflicts and let me back up and confirm replacement before --force.
7. Tell me to exit this session and run erp-skill launch claude-code in a terminal. Bare claude can load persistent configuration but does not automatically fetch the ERP credential from this project's keyring. Do not nest interactive sessions.
8. Report paths, version, actual checks, and remaining steps. Do not claim completion while credential input or restarting is still pending.
```

## WorkBuddy

```text
Prepare an ERP Agent Skill Connector for WorkBuddy for the current user, then guide me through importing it.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.1
MCP URL: https://erp-open.example.com/mcp (ask me to replace this placeholder before execution)

1. Check OS, architecture, terminal permission, and network access. Prefer prebuilt Release packages over source compilation.
2. Download the matching archive and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1, verify SHA-256, install erp-skill, configure PATH, and confirm version 0.1.1.
3. Run:
   erp-skill install workbuddy --url <real-MCP-URL> --output ./erp-open-platform-connector
   Report the absolute output directory. Explain conflicts and let me back up and confirm before --force.
4. Generation requires no personal Key. Do not ask for a Key in chat or write a literal Key into JSON. Preserve ${ERP_OPEN_API_KEY}.
5. Guide me to import the generated directory in WorkBuddy's Connector UI, enter the Key in its Token UI, and connect or reconnect MCP. The downloaded CLI archive is not itself an importable Connector.
6. Explicitly list manual UI steps if the current client cannot perform them. Do not use erp-skill launch workbuddy; it is unsupported.
7. If I want independent CLI connectivity verification, provide hidden-input code for erp-skill doctor workbuddy --url <real-MCP-URL> --key-stdin. Explain that it does not populate WorkBuddy's Token UI.
8. Report the executable version, Connector path, actual check results, and UI connection status. A generated directory alone is not verified integration.
```

## Completion criteria

- The executable version matches the selected release.
- Codex / Claude Code installation succeeds, `doctor` reports its actual result, and a new session starts with `erp-skill launch`.
- WorkBuddy has a generated Connector, completed UI import, configured Token, and an active MCP connection. Standalone `doctor` only verifies the endpoint and Key supplied to it.
- `INSTALLED_AUTHORIZATION_REQUIRED` is not equivalent to full `VERIFIED` status.

See the [installation guide](installation.en.md) for complete hidden-input commands and the distinction between initialization, installation, and authenticated startup.
