# Installation, upgrades, and troubleshooting

[中文](installation.md) | English

## Choose an installation method

Use the [prebuilt Release packages](https://github.com/xsy-tech/erp-agent-skills/releases) for ordinary installations. You do not need Go, GitHub CLI, repository write access, or source compilation. Install Codex, Claude Code, or WorkBuddy separately; `erp-skill` configures and launches integrations, rather than replacing an agent client.

Obtain the MCP URL and a personal API Key bound to the correct ERP user and tenant. Codex / Claude Code require an available OS credential manager: macOS Keychain, Windows Credential Manager, or Linux Secret Service. A headless Linux host also needs a working Secret Service backend; the installer does not fall back to plaintext storage.

| OS | Processor | v0.1.1 package |
| --- | --- | --- |
| macOS | Intel | `erp-agent-skill_0.1.1_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `erp-agent-skill_0.1.1_darwin_arm64.tar.gz` |
| Linux | Intel / AMD 64-bit | `erp-agent-skill_0.1.1_linux_amd64.tar.gz` |
| Linux | ARM 64-bit | `erp-agent-skill_0.1.1_linux_arm64.tar.gz` |
| Windows | x64 | `erp-agent-skill_0.1.1_windows_amd64.zip` |

These packages contain the CLI; they are not directly importable WorkBuddy Connectors. Generate a Connector after installing the CLI.

## One-command installation of the published version

### macOS / Linux

Run in Bash or Zsh:

```bash
set -o pipefail && curl -fsSL https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.sh | ERP_SKILL_RELEASE_BASE_URL=https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1 bash && export PATH="$HOME/.local/bin:$PATH" && erp-skill version
```

The pinned script detects the OS and architecture, downloads the archive and `checksums.txt`, and installs only after verification. It requires Bash, `curl`, `tar`, and `shasum`; Perl commonly provides `shasum` on Linux. If you have only `sha256sum`, use the newer repository script or the manual procedure below.

For future terminals, add this to your `~/.zshrc` or `~/.bashrc` and reopen the terminal:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Windows x64

Run in PowerShell:

```powershell
$ErrorActionPreference = 'Stop'; & ([scriptblock]::Create((Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/xsy-tech/erp-agent-skills/v0.1.1/scripts/install.ps1').Content)) -Version '0.1.1' -BaseUrl 'https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1'; $env:Path = "$HOME\.local\bin;$env:Path"; erp-skill version
```

The script verifies SHA-256 before installing into `$HOME\.local\bin`. Add that directory to your user Path in Windows Environment Variables and reopen the terminal for persistent access.

### Newer repository scripts

After downloading or checking out the repository:

```bash
bash scripts/install.sh
# Explicitly choose an already published version:
ERP_SKILL_VERSION=0.1.1 bash scripts/install.sh
# Organization mirror and custom destination:
ERP_SKILL_VERSION=0.1.1 ERP_SKILL_RELEASE_BASE_URL=https://downloads.example.com/erp/v0.1.1 ERP_SKILL_INSTALL_DIR="$HOME/.local/bin" bash scripts/install.sh
```

```powershell
& .\scripts\install.ps1
& .\scripts\install.ps1 -Version '0.1.1' -BaseUrl 'https://downloads.example.com/erp/v0.1.1' -InstallDir "$HOME\.local\bin"
```

The newer scripts default to the official GitHub Release URL. The shell script prefers `sha256sum`, with `shasum` as a fallback. Replace example mirror URLs with real ones; archive versions and checksum manifests must agree.

## Manual installation

Download the matching archive and `checksums.txt` from [v0.1.1 Assets](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.1).

For Apple Silicon, calculate the hash and compare it with the matching manifest entry:

```bash
shasum -a 256 erp-agent-skill_0.1.1_darwin_arm64.tar.gz
```

Linux can use `sha256sum <archive>`. Do not check the entire manifest when you have downloaded only one archive; the manifest lists other platforms too. After the matching hash is verified:

```bash
tar -xzf erp-agent-skill_0.1.1_darwin_arm64.tar.gz
mkdir -p "$HOME/.local/bin"
install -m 0755 erp-agent-skill_0.1.1_darwin_arm64/erp-skill "$HOME/.local/bin/erp-skill"
export PATH="$HOME/.local/bin:$PATH"
erp-skill version
```

On Windows, use `Get-FileHash <archive.zip> -Algorithm SHA256`, compare the digest, expand with `Expand-Archive`, and place `erp-skill.exe` in a directory on your user PATH.

## Connect Codex or Claude Code

### macOS / Linux (Bash)

Replace the URL; select `codex` or `claude-code`:

```bash
MCP_URL='https://erp-open.example.com/mcp'
AGENT=codex
IFS= read -r -s -p 'Personal ERP API Key: ' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install "$AGENT" --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor "$AGENT" --url "$MCP_URL"
```

`read -p` is Bash syntax. If your interactive shell is Zsh, enter `bash` before pasting this block.

### Windows (PowerShell)

```powershell
$McpUrl = 'https://erp-open.example.com/mcp'
$Agent = 'codex' # Use claude-code for Claude Code
$SecureKey = Read-Host 'Personal ERP API Key' -AsSecureString
$KeyPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureKey)
try {
    $PlainKey = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($KeyPointer)
    $PlainKey | erp-skill install $Agent --url $McpUrl --key-stdin
    if ($LASTEXITCODE -ne 0) { throw 'Agent installation failed; inspect the error.' }
} finally {
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($KeyPointer)
    Remove-Variable PlainKey, SecureKey -ErrorAction SilentlyContinue
}
erp-skill doctor $Agent --url $McpUrl
```

The Key is entered through hidden input and passed through stdin, not literal command arguments, chat, or configuration. Plaintext exists briefly in process memory for transmission; the installer stores it in the OS credential manager.

### Start a new session

Exit the current agent session, then run one of:

```bash
erp-skill launch codex
erp-skill launch claude-code
```

Do not ask an already-running agent to nest a new interactive session. Installation persists the shared Skill, references, and the `erp-open` MCP registration. Start a new session or reconnect and check the current `tools/list`.

## Starting codex or claude directly

An MCP `initialize` handshake checks a connection; it does not install a Skill or persist client configuration. The following applies after successful `erp-skill install codex` or `erp-skill install claude-code`:

| Item | New bare codex / claude process | erp-skill launch |
| --- | --- | --- |
| Installed ERP Skill | Discoverable from the client's local Skill directory; full instructions are used when relevant. | Uses the same installed Skill. |
| Persisted erp-open MCP configuration | Read by the configured client, subject to its settings and version. | Uses the same registration. |
| Key stored by erp-skill in the OS credential manager | This project's keyring loader is not invoked. | Read and injected into the child agent process. |
| Successful authenticated ERP calls | Requires the matching credential environment variable already present and valid. | Uses the stored credential; endpoint access and server authorization still apply. |

The required variables are `ERP_OPEN_CODEX_KEY` for Codex and `ERP_OPEN_CLAUDE_KEY` for Claude Code. Saving a Key during installation does not export it into the parent terminal; running `launch` once does not authenticate separate terminals or a later unrelated `codex` / `claude` process. Successful `doctor` checks also do not export these variables.

The current installer places Codex Skills in `$CODEX_HOME/skills/erp-open-platform` (default `~/.codex/skills/erp-open-platform`) and Claude Code Skills in `~/.claude/skills/erp-open-platform`. Skill discovery locations depend on the client version; check the installed client's Skill list if it does not appear. Do not assume this covers every future client release.

For routine use, prefer `launch`. Do not add literal Keys to shell startup files or MCP configuration. There is no automatic shell alias or global credential export installed by this project.

Sources: [Codex MCP](https://developers.openai.com/codex/mcp/), [Codex Skills](https://developers.openai.com/codex/skills/), [Claude Code MCP](https://code.claude.com/docs/en/mcp), and [Claude Code Skills](https://code.claude.com/docs/en/skills). The paths and keyring behavior above describe this repository's implementation.

## Connect WorkBuddy

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

Import the generated directory through WorkBuddy's Connector UI, enter your personal Key in the Token UI, and connect or reconnect MCP. Generation does not need a Key; keep `${ERP_OPEN_API_KEY}` as a placeholder in files. WorkBuddy does not use `erp-skill launch`.

For an independent connectivity check, use hidden input as above and replace the install command with:

```bash
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill doctor workbuddy --url "$MCP_URL" --key-stdin
```

Then run `unset ERP_INSTALL_KEY`. In PowerShell, replace `install $Agent` with `doctor workbuddy` in the hidden-input block. This does not configure the WorkBuddy Token UI.

## Upgrade, roll back, or uninstall

Select an already published target version, reinstall the matching CLI archive using its script and checksum manifest, then check `erp-skill version`.

An updated Skill can conflict with files already on disk. Back up local changes, inspect them, and add `--force` to the secure `install` command only after deciding to replace them. For WorkBuddy, regenerate the Connector with `--force` after confirming replacement, then reimport it.

```bash
erp-skill uninstall codex --url https://erp-open.example.com/mcp
erp-skill uninstall claude-code --url https://erp-open.example.com/mcp
```

Remove WorkBuddy Connectors and Tokens through its UI. Once no client uses the CLI, remove `~/.local/bin/erp-skill` or the Windows executable manually. Local removal does not revoke a server-side Key; ask an administrator to revoke or rotate compromised credentials.

## Troubleshooting

| Symptom | Action |
| --- | --- |
| Command not found | Add the installation directory to PATH or use its absolute path. |
| Download returns 404 | Check the published version, platform, filename, and mirror. |
| Checksum mismatch | Stop; redownload the matching archive and manifest. Do not bypass verification. |
| shasum missing on Linux | The pinned v0.1.1 script needs it; install Perl, verify manually, or use the newer repository script. |
| Credential manager unavailable | Repair Keychain / Credential Manager / Secret Service; do not use plaintext fallback storage. |
| Existing files differ | Back them up and inspect them before using --force. |
| INSTALLED_AUTHORIZATION_REQUIRED | Ask an administrator to authorize read-only get_platform_info. |
| Bare codex / claude sees ERP configuration but cannot call tools | Prefer erp-skill launch; bare startup does not load this project's keyring credential. |
| Added tools are missing | Reconnect MCP or start a new session and inspect tools/list. |
| Only a WorkBuddy directory was generated | Complete Connector import, Token configuration, and connection in WorkBuddy. |

See [installation prompts](agent-install-prompts.en.md), [agent usage](agent-usage.en.md), and [error handling](../references/errors.en.md).
