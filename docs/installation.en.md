# Installation, upgrades, and troubleshooting

[中文](installation.md) | English

> Version note: persistent authentication and direct startup below describe the v0.1.2 installer (not yet published). Published v0.1.1 still uses the legacy environment-variable flow. Install a release containing this change, or build the newer source and reinstall the integration. Existing release assets are not replaced.

## Choose an installation method

Use the [prebuilt Release packages](https://github.com/xsy-tech/erp-agent-skills/releases) for ordinary installations. You do not need Go, GitHub CLI, repository write access, or source compilation. Install Codex, Claude Code, or WorkBuddy separately; `erp-skill` configures and launches integrations, rather than replacing an agent client.

Obtain the MCP URL and a personal API Key bound to the correct ERP user and tenant. The newer installer persists authentication in private user-level Codex / Claude Code configuration. It no longer requires an OS keyring or Linux Secret Service. Old keyring entries are read only for compatibility when a persistent credential is unavailable.

| OS | Processor | v0.1.1 package |
| --- | --- | --- |
| macOS | Intel | `erp-agent-skill_0.1.1_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `erp-agent-skill_0.1.1_darwin_arm64.tar.gz` |
| Linux | Intel / AMD 64-bit | `erp-agent-skill_0.1.1_linux_amd64.tar.gz` |
| Linux | ARM 64-bit | `erp-agent-skill_0.1.1_linux_arm64.tar.gz` |
| Windows | x64 | `erp-agent-skill_0.1.1_windows_amd64.zip` |

These packages contain the CLI; they are not directly importable WorkBuddy Connectors. The newer installer configures WorkBuddy MCP and Skill directly; generating a Connector is needed only for the optional distribution flow.

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

The Key is entered through hidden input and passed through stdin, never literal command arguments or chat. The installer saves it in the current user’s private MCP configuration for direct client access. Do not commit or share that file.

### Start a new session

Restart the client after installation, then use the usual commands:

```bash
codex
# Or:
claude
```

Local desktop Code sessions using the same user configuration can also connect directly. Reconnect MCP or restart an already-running client to reload its configuration. No ERP credential environment variable or nested interactive session is required.

## Starting codex or claude directly

The newer installer persists the shared Skill, URL, and authentication header:

| Client | User configuration | Credential field |
| --- | --- | --- |
| Codex CLI / local Codex desktop using the same configuration | $CODEX_HOME/config.toml; default ~/.codex/config.toml | mcp_servers.erp-open.http_headers.Authorization |
| Claude Code CLI / local desktop Code session | ~/.claude.json; with CLAUDE_CONFIG_DIR, .claude.json inside that directory | mcpServers.erp-open.headers.Authorization |

Codex configuration uses this form; the example Key is a placeholder, and the installer receives the real value through stdin:

```toml
[mcp_servers.erp-open]
url = "https://erp-open.example.com/mcp"
[mcp_servers.erp-open.http_headers]
Authorization = "Bearer <YOUR_PERSONAL_KEY>"
```

Claude Code uses:

```json
{
  "mcpServers": {
    "erp-open": {
      "type": "http",
      "url": "https://erp-open.example.com/mcp",
      "headers": { "Authorization": "Bearer <YOUR_PERSONAL_KEY>" }
    }
  }
}
```

These are private user files, not a project .mcp.json or files to commit. Unix permissions are set to 0600; Windows uses the user profile's access controls. Updates change erp-open while preserving other servers and settings. Keys are never passed through client command arguments or printed. Invalid existing configurations are rejected without overwriting the file.

After restarting under the same account, the client can read the saved credential. Skill instructions remain loaded on demand for relevant work. An initialize handshake alone does not install a Skill.

The current Skill paths are $CODEX_HOME/skills/erp-open-platform (default ~/.codex/skills/erp-open-platform) and ~/.claude/skills/erp-open-platform. With CLAUDE_CONFIG_DIR, Claude uses skills/erp-open-platform inside that directory. Client versions and disabling policies can affect discovery.

`erp-skill launch` remains a compatibility option. Both doctor and launch read persistent configuration first, falling back to an old OS credential only when necessary. New installations require no OS keyring.

**Migration:** published v0.1.1 still uses ERP_OPEN_CODEX_KEY / ${ERP_OPEN_CLAUDE_KEY} and keyring credentials, so its launcher flow remains necessary. Install a newer release containing this change, rerun the secure install command, and replace conflicting Skill files with --force only after backup and confirmation. The installer replaces the old ERP environment-variable registration with fixed headers; then restart the client.

**Desktop scope:** this describes local Codex and Claude Code Code sessions reading these user files. Ordinary Claude Desktop chat and cloud sessions have different configuration entry points and require their own import or authorization; they are not claimed to work merely because a local Code integration was installed.

Sources: [Codex MCP](https://developers.openai.com/codex/mcp/), [Codex configuration](https://developers.openai.com/codex/config-reference/), [Claude Code MCP](https://code.claude.com/docs/en/mcp), and [Claude Code desktop](https://code.claude.com/docs/en/desktop).

## Connect WorkBuddy

### Default: install MCP and Skill directly (v0.1.2)

This feature belongs to v0.1.2, which is not yet published. Published v0.1.1 exports Connectors only. Use a release containing this feature, or build the newer source before running these commands.

In Bash (enter bash first if your shell is Zsh), replace the endpoint:

```bash
MCP_URL='https://erp-open.example.com/mcp'
IFS= read -r -s -p 'Personal ERP API Key: ' ERP_INSTALL_KEY; printf '\n'
printf '%s\n' "$ERP_INSTALL_KEY" | erp-skill install workbuddy --url "$MCP_URL" --key-stdin
unset ERP_INSTALL_KEY
erp-skill doctor workbuddy --url "$MCP_URL"
```

For Windows, use the PowerShell hidden-input block above with `$Agent = 'workbuddy'`. The install command remains `erp-skill install $Agent --url $McpUrl --key-stdin`; then run `erp-skill doctor workbuddy --url $McpUrl`.

The installer performs read-only MCP checks and installs user-level configuration and Skill:

| Item | Path or field |
| --- | --- |
| User MCP configuration | ~/.workbuddy/mcp.json; on Windows, %USERPROFILE%\.workbuddy\mcp.json |
| ERP server | mcpServers.erp-open |
| Transport | streamableHttp |
| Credential | headers.Authorization containing Bearer <personal-Key> |
| ERP Skill | ~/.workbuddy/skills/erp-open-platform/SKILL.md and bilingual references |

Updates preserve unrelated servers, settings, and JSON numbers. Invalid existing configuration is rejected. The Key is saved only in private user configuration, never process arguments or chat; Unix mode is 0600. Direct installation needs no Connector import, Token form, or launcher.

Restart WorkBuddy or reconnect erp-open in its custom MCP management UI and confirm the Skill is enabled. Check for same-named project-level MCP or Skill overrides if needed; the installer modifies only user-level files.

Doctor reads the saved credential by default. Optional `--key-stdin` independently verifies a supplied Key without replacing the stored credential.

### Optional: export a Connector

Use this only to distribute a Connector with a Token form, icon, and Skill:

```bash
erp-skill install workbuddy --url https://erp-open.example.com/mcp --output ./erp-open-platform-connector
```

Export creates files only, without registering local MCP, saving a Key, or installing a local Skill. It requires no Key and cannot be combined with --key-stdin. Files retain ${ERP_OPEN_API_KEY}; import through WorkBuddy and enter the Key in its Token form. Legacy v0.1.1 users can still use this flow.

### Migrate from an imported Connector

Disable the old ERP Connector first and keep its files. Avoid connecting both that Connector and the new erp-open registration, which would duplicate tools. Install and verify with the newer CLI, restart WorkBuddy, and inspect the MCP and Skill. Remove the old Connector through the UI after confirming the new connection works. The installer does not delete imported Connectors.

The [official WorkBuddy MCP guide](https://www.codebuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/MCP-Guide) describes user configuration paths. File installation and protocol checks do not constitute actual WorkBuddy UI acceptance testing.

## Upgrade, roll back, or uninstall

Select an already published target version, reinstall the matching CLI archive using its script and checksum manifest, then check `erp-skill version`.

An updated Skill can conflict with files already on disk. Back up local changes, inspect them, and add `--force` to the secure `install` command only after deciding to replace them. For direct WorkBuddy installation, rerun the secure install command with --force after backup and confirmation. Only the optional Connector flow requires regeneration and reimport.

```bash
erp-skill uninstall codex --url https://erp-open.example.com/mcp
erp-skill uninstall claude-code --url https://erp-open.example.com/mcp
```

For direct installation, run `erp-skill uninstall workbuddy --url <real-MCP-URL>` to remove the local ERP credential and unchanged managed Skill while preserving other settings. Optional Connectors and Tokens are still removed through WorkBuddy’s UI. Once no client uses the CLI, remove `~/.local/bin/erp-skill` or the Windows executable manually. Local removal does not revoke a server-side Key; ask an administrator to revoke or rotate compromised credentials.

## Troubleshooting

| Symptom | Action |
| --- | --- |
| Command not found | Add the installation directory to PATH or use its absolute path. |
| Download returns 404 | Check the published version, platform, filename, and mirror. |
| Checksum mismatch | Stop; redownload the matching archive and manifest. Do not bypass verification. |
| shasum missing on Linux | The pinned v0.1.1 script needs it; install Perl, verify manually, or use the newer repository script. |
| Legacy credential manager unavailable | Repair it for v0.1.1, or migrate to the newer installer, which uses private user configuration. |
| Existing files differ | Back them up and inspect them before using --force. |
| INSTALLED_AUTHORIZATION_REQUIRED | Ask an administrator to authorize read-only get_platform_info. |
| Direct client startup cannot call ERP | Confirm migration and restart, run doctor, check Key validity, and check for project-level overrides. |
| Added tools are missing | Reconnect MCP or start a new session and inspect tools/list. |
| Only a WorkBuddy directory was generated | You selected optional --output export; omit --output and use --key-stdin for direct installation. |

See [installation prompts](agent-install-prompts.en.md), [agent usage](agent-usage.en.md), and [error handling](../references/errors.en.md).
