# ERP Agent Skill

[中文](README.md) | English

> v0.1.2 supports persistent authentication: install the integration, restart your client, and use it normally. Upgrade and reinstall integrations from legacy v0.1.1; existing release assets remain unchanged.

Connect Codex, Claude Code, and WorkBuddy to the ERP Open Platform MCP server using the capabilities authorized for your personal ERP identity.

**Install a prebuilt GitHub Release package first. You do not need Go, a source checkout, or a local build.** The published version is [v0.1.2](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2).

## One-command installation

These commands download, verify, and install `erp-skill`. Install your chosen agent separately, then obtain the MCP URL and a personal API Key from your administrator.

### macOS / Linux (Bash or Zsh)

```bash
set -o pipefail && curl -fsSL https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.2/scripts/install.sh | ERP_SKILL_RELEASE_BASE_URL=https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.2 bash && export PATH="$HOME/.local/bin:$PATH" && erp-skill version
```

The installer supports either `sha256sum` or `shasum`. See the [installation guide](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md).

### Windows x64 (PowerShell)

```powershell
$ErrorActionPreference = 'Stop'; & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.2/scripts/install.ps1').Content)) -Version '0.1.2' -BaseUrl 'https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.2'; $env:Path = "$HOME\.local\bin;$env:Path"; erp-skill version
```

Both commands install into the current user's `~/.local/bin` without administrator privileges and update PATH only for the current terminal. See the installation guide for persistent PATH settings, manual downloads, upgrades, and troubleshooting.

## Ask your agent to install it

Copy the relevant [installation prompt](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-install-prompts.en.md) into Codex, Claude Code, or WorkBuddy. The prompts prefer prebuilt packages and require SHA-256 verification and truthful completion reporting.

Do not paste your personal API Key into chat. Use hidden terminal input for direct installation into all three clients. WorkBuddy's Token UI applies only to optional Connector packages.

## Connect Codex / Claude Code

In Bash, replace the MCP URL and choose `codex` or `claude-code`:

```bash
MCP_URL='https://erp-open.example.com/mcp'
AGENT=codex # Use claude-code for Claude Code
IFS= read -r -s -p 'Personal ERP API Key: ' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install "$AGENT" --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor "$AGENT" --url "$MCP_URL"
```

After installation, restart the client, then use the usual commands:

```bash
codex
# Or:
claude
```

You can also open a local Codex desktop session or Claude Code desktop Code session that reads the same user configuration. The installer persists the MCP URL and `Authorization: Bearer …` header alongside the shared Skill. No special launcher or shell credential variable is required.

### Can I run codex or claude directly?

Yes, after installing with the newer installer. The client reads the persistent `erp-open` connection and credential and discovers the ERP Skill. Full Skill instructions are loaded when relevant, rather than applied to every task.

Codex uses `$CODEX_HOME/config.toml` (default `~/.codex/config.toml`); Claude Code uses user-scoped `~/.claude.json`. These files contain the personal Key, stay under the current user's profile, and are written with Unix mode `0600`. Do not commit or share them. The installer never passes a Key through client command-line arguments.

`erp-skill launch` remains a compatibility option. `doctor` reads the same persistent credential first, with legacy keyring credentials as a fallback. An MCP `initialize` handshake alone does not install a Skill. Ordinary Claude Desktop chat and cloud sessions use different configuration entry points; this does not claim those are the same as local Claude Code sessions. See the [startup guide](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md#starting-codex-or-claude-directly).

## Connect WorkBuddy

**v0.1.2 installs MCP and the Skill directly by default; no Connector-package import is required.** Legacy v0.1.1 supports only the optional Connector-export flow below.

In Bash, enter the personal Key through hidden input and replace the MCP URL:

```bash
MCP_URL='https://erp-open.example.com/mcp'
IFS= read -r -s -p 'Personal ERP API Key: ' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install workbuddy --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor workbuddy --url "$MCP_URL"
```

The installer updates erp-open in `~/.workbuddy/mcp.json` with the streamableHttp URL and Authorization header and installs the Skill into `~/.workbuddy/skills/erp-open-platform`. Restart WorkBuddy or reconnect the server and confirm the Skill is enabled. No launcher or separate Connector Token form is needed. Other MCP settings are preserved; the configuration contains the personal Key and uses Unix mode 0600.

For optional Connector distribution instead of local installation:

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

`--output` exports a package without a real Key; import it and fill in the Token through WorkBuddy. It cannot be combined with `--key-stdin`. See the [installation guide](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md#connect-workbuddy) for Windows input, migration, and removal.

## Use ERP tools

Describe business goals in Chinese or English:

- "Find employees whose name contains Zhang and show their employee numbers and departments."
- "Read the order's current state before explaining whether it can be submitted."
- "Show the customer, products, quantities, and amounts before creating the sales document; wait for my confirmation."

The current MCP `tools/list` is authoritative. The server determines your ERP user, tenant, and data scope. Reconnect after authorization changes.

`VERIFIED` means the read-only installation checks succeeded. `INSTALLED_AUTHORIZATION_REQUIRED` means initialization and discovery succeeded but `get_platform_info` is not authorized; ask an administrator to grant it. Verification does not invoke write-capable business tools.

## English documentation

- [v0.1.2 project1 verification and known blockers](docs/verification-v0.1.2-project1.en.md)
- [Installation, upgrades, removal, and troubleshooting](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md)
- [Installation prompts for all three agents](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-install-prompts.en.md)
- [Agent usage](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-usage.en.md)
- [Build and release operations](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/enterprise-release.en.md)
- [Bilingual shared Skill](https://github.com/xsy-tech/erp-agent-skills/blob/main/SKILL.md)
- [Terminology](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/terminology.en.md), [workflows](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/workflows.en.md), [safety](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/safety.en.md), and [errors](https://github.com/xsy-tech/erp-agent-skills/blob/main/references/errors.en.md)

## Developers: source builds (alternative)

Use the Go version declared in `go.mod` when modifying this project:

```bash
git clone https://github.com/xsy-tech/erp-agent-skills.git
cd erp-agent-skills
make check
make build
./bin/erp-skill version
```

Maintainers use `make dist` for all five platform packages and `make release` to publish a new GitHub Release. See the release guide.
