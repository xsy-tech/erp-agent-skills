# ERP Agent Skill

[中文](README.md) | English

Connect Codex, Claude Code, and WorkBuddy to the ERP Open Platform MCP server using the capabilities authorized for your personal ERP identity.

**Install a prebuilt GitHub Release package first. You do not need Go, a source checkout, or a local build.** The published version is [v0.1.1](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1).

## One-command installation

These commands download, verify, and install `erp-skill`. Install your chosen agent separately, then obtain the MCP URL and a personal API Key from your administrator.

### macOS / Linux (Bash or Zsh)

```bash
set -o pipefail && curl -fsSL https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.sh | ERP_SKILL_RELEASE_BASE_URL=https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1 bash && export PATH="$HOME/.local/bin:$PATH" && erp-skill version
```

The pinned v0.1.1 script requires `shasum`, usually provided by Perl on Linux. The newer repository script also supports `sha256sum`. See the [installation guide](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md).

### Windows x64 (PowerShell)

```powershell
$ErrorActionPreference = 'Stop'; & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.ps1').Content)) -Version '0.1.1' -BaseUrl 'https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1'; $env:Path = "$HOME\.local\bin;$env:Path"; erp-skill version
```

Both commands install into the current user's `~/.local/bin` without administrator privileges and update PATH only for the current terminal. See the installation guide for persistent PATH settings, manual downloads, upgrades, and troubleshooting.

## Ask your agent to install it

Copy the relevant [installation prompt](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/agent-install-prompts.en.md) into Codex, Claude Code, or WorkBuddy. The prompts prefer prebuilt packages and require SHA-256 verification and truthful completion reporting.

Do not paste your personal API Key into chat. Enter it through hidden terminal input for Codex / Claude Code, or through WorkBuddy's Token UI.

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

After installation, exit the existing agent session and start a new one:

```bash
erp-skill launch codex
# Or:
erp-skill launch claude-code
```

The installer persists the Skill and `erp-open` MCP registration and stores your Key in the OS credential manager. `launch` reads that Key and injects it into the child agent process.

### Can I run codex or claude directly?

After `erp-skill install`, a normal new local agent session can discover the installed Skill and MCP configuration. Skill instructions are loaded when relevant to a request, rather than all being applied to every task.

**Configuration discovery does not supply credentials.** Bare `codex` or `claude` does not invoke this project's keyring loader. Direct startup can authenticate only if its process already receives `ERP_OPEN_CODEX_KEY` or `ERP_OPEN_CLAUDE_KEY`, respectively. This installer does not export those variables to your parent shell. Prefer `erp-skill launch` instead of storing plaintext keys in shell profiles.

An MCP `initialize` handshake by itself does not install a Skill or persist a client registration. See the [startup FAQ](https://github.com/xsy-tech/erp-agent-skills/blob/main/docs/installation.en.md#starting-codex-or-claude-directly).

## Connect WorkBuddy

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

Import the generated directory into WorkBuddy, enter your Key in its Token UI, and connect or reconnect MCP. Generated files retain `${ERP_OPEN_API_KEY}` as a placeholder. `doctor workbuddy` can check connectivity independently but does not import the connector or configure its Token. WorkBuddy does not use `erp-skill launch`.

## Use ERP tools

Describe business goals in Chinese or English:

- "Find employees whose name contains Zhang and show their employee numbers and departments."
- "Read the order's current state before explaining whether it can be submitted."
- "Show the customer, products, quantities, and amounts before creating the sales document; wait for my confirmation."

The current MCP `tools/list` is authoritative. The server determines your ERP user, tenant, and data scope. Reconnect after authorization changes.

`VERIFIED` means the read-only installation checks succeeded. `INSTALLED_AUTHORIZATION_REQUIRED` means initialization and discovery succeeded but `get_platform_info` is not authorized; ask an administrator to grant it. Verification does not invoke write-capable business tools.

## English documentation

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
