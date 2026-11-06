# Installation prompts for your agent

[中文](agent-install-prompts.md) | English

Paste the relevant prompt into Codex, Claude Code, or WorkBuddy. Replace the MCP URL before execution. Never paste a personal API Key into the prompt.

An agent needs terminal, network, and filesystem access to perform these steps. Without that access, it should provide local commands and report the remaining actions. Hidden credential entry, client UI import, and restarting an interactive session may require the user.

## Codex

```text
Install ERP Agent Skill for the current user and configure Codex on this computer.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.3
MCP URL: https://erp-open.example.com/mcp (ask me to replace this placeholder before execution)

1. Check OS, processor architecture, network access, and the codex executable. Prefer the matching prebuilt Release archive. Do not install Go or compile source by default.
2. Download the matching archive and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.3. Verify SHA-256 before installation. Alternatively, use that Tag's scripts/install.sh or scripts/install.ps1 with the base URL https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.3, after checking its dependencies. Stop on checksum failure.
3. Add the current user's installation directory to PATH and verify erp-skill version reports 0.1.3.
4. Do not request a Key in chat or place it in shared files or literal command arguments. The newer installer writes it into private user-level MCP configuration. Give me local hidden-input code that passes it through stdin to:
   erp-skill install codex --url <real-MCP-URL> --key-stdin
5. Run erp-skill doctor codex --url <real-MCP-URL> after installation, using only connection and read-only verification.
6. If managed files conflict, explain the affected files and ask me to back them up and confirm replacement before using --force.
7. Check the installed version: with the newer persistent-auth installer, restart the client and use codex or a matching local desktop Code session directly. Published v0.1.1 still needs erp-skill launch codex. Do not claim legacy packages have the new behavior or nest interactive sessions.
8. Report the installation path, version, actual verification result, and remaining user steps. Missing credential input or a required restart must not be reported as verified success.
```

After setup, ask: "List the read-only capabilities in the current ERP MCP connection and explain my identity and data scope. Do not write business data."

## Claude Code

```text
Install ERP Agent Skill for the current user and configure Claude Code on this computer.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.3
MCP URL: https://erp-open.example.com/mcp (ask me to replace this placeholder before execution)

1. Check OS, architecture, network access, and the claude executable. Prefer a prebuilt Release archive over a source build.
2. Download the matching package and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.3. Verify SHA-256. The corresponding Tag's installer script can also be used with base URL https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.3. Stop if verification fails.
3. Set the user's PATH and confirm erp-skill version reports 0.1.3.
4. Do not collect a Key in chat. Provide local hidden-input code that passes the Key through stdin to:
   erp-skill install claude-code --url <real-MCP-URL> --key-stdin
   Do not put Keys in chat, process arguments, project configuration, or the repository. The newer installer saves Authorization in private user configuration. Legacy v0.1.1 still uses the ${ERP_OPEN_CLAUDE_KEY} placeholder; distinguish these versions.
5. Run erp-skill doctor claude-code --url <real-MCP-URL> using read-only checks.
6. Explain existing-file conflicts and let me back up and confirm replacement before --force.
7. With the newer persistent-auth installer, restart and use claude or a local Claude Code desktop Code session directly. Legacy v0.1.1 still needs erp-skill launch claude-code. Do not nest interactive sessions.
8. Report paths, version, actual checks, and remaining steps. Do not claim completion while credential input or restarting is still pending.
```

## WorkBuddy

Use the v0.1.3 prebuilt package for direct installation. Legacy v0.1.1 supports Connector export only.

```text
Install ERP MCP and the ERP Skill directly for the current WorkBuddy user. Do not require Connector-package import.

Official repository: https://github.com/xsy-tech/erp-agent-skills
Release: v0.1.3. Do not silently substitute v0.1.1, which exports Connectors only.
MCP URL: https://erp-open.example.com/mcp (ask me to replace the placeholder before execution)

1. Check OS, architecture, terminal/network access, and the installed WorkBuddy client. Download the matching prebuilt package and checksums.txt from https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.3; verify SHA-256 and erp-skill version reports 0.1.3. Stop if download or verification fails; do not install Go or compile by default.
2. Do not request a Key in chat. Provide local hidden-input code that pipes it to:
   erp-skill install workbuddy --url <real-MCP-URL> --key-stdin
   Do not add --output or require a Connector Token form for this mode.
3. The installer updates erp-open in ~/.workbuddy/mcp.json using streamableHttp and headers.Authorization, and installs ~/.workbuddy/skills/erp-open-platform. This is private user configuration containing a Key; never put it in a shared project, repository, or chat, and do not print its contents.
4. Preserve unrelated MCP servers/settings. Explain Skill conflicts and let me back up and confirm before --force.
5. Run erp-skill doctor workbuddy --url <real-MCP-URL> using the saved credential and read-only verification. Add --key-stdin only when I explicitly want an independent check of another Key.
6. If an old ERP Connector is enabled, guide me to disable it to avoid duplicate tools. Do not automatically delete it.
7. Tell me to restart WorkBuddy or reconnect erp-open in its MCP UI and confirm the Skill is enabled. Do not use erp-skill launch workbuddy.
8. Report version, paths, actual doctor status, and remaining reconnect steps. File installation and protocol checks are not actual WorkBuddy UI acceptance testing.
```

For optional package distribution, use `erp-skill install workbuddy --url <MCP-URL> --output ./erp-open-platform-connector`. Export accepts no Key and does not register local MCP or install a local Skill; import the package and fill its Token form instead. Choose one flow.

## Completion criteria

- The executable version matches the selected release.
- Codex / Claude Code installation succeeds, `doctor` reports its actual result, and the client restarts using the installed version’s authentication flow: direct startup for the newer installer, or the legacy launcher for v0.1.1.
- WorkBuddy direct installation writes user MCP configuration and Skill, verifies the saved credential with doctor, and requires client reconnection to confirm MCP/Skill availability. Optional Connector export separately needs UI import and Token configuration.
- `INSTALLED_AUTHORIZATION_REQUIRED` is not equivalent to full `VERIFIED` status.

See the [installation guide](installation.en.md) for complete hidden-input commands and the distinction between initialization, installation, and authenticated startup.
