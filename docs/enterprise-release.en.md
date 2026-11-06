# Builds, enterprise distribution, and GitHub releases

[中文](enterprise-release.md) | English

This guide is for maintainers. Ordinary users should follow the [prebuilt installation guide](installation.en.md), without Go or a local build.

## Repository and access

The default release repository is `xsy-tech/erp-agent-skills`. Keep repository write and tag access limited to maintainers. Users only need immutable versioned downloads.

## Source development and builds

Use the Go version declared in `go.mod`:

```bash
git clone https://github.com/xsy-tech/erp-agent-skills.git
cd erp-agent-skills
make check
make build
./bin/erp-skill version
```

`make build` creates a native CLI. `make dist` builds five platform archives and `dist/checksums.txt`. Both `README.md` and `README.en.md` belong in every archive. The installed Skill and WorkBuddy Connector must include the Chinese and English reference files. `make help` lists available targets.

## One-command release

```bash
make release
# Or explicitly assert the expected source version:
make release VERSION=0.1.2
```

Before publishing:

1. Update `version.go`, `manifest.json`, `adapters/workbuddy/connector-meta.json`, `scripts/install.sh`, and `scripts/install.ps1` together. Update pinned versions and URLs in both languages of the README, installation guide, and prompts.
2. Run `make release-check`: tests, `go vet`, five platform builds, and SHA-256 verification. It does not create tags or releases.
3. Review source and artifacts for credentials, then commit and push all release changes. Publishing requires a clean working tree, including untracked files.
4. Install GitHub CLI and authenticate once with `gh auth login` using an account with repository write access.

`make release` repeats checks, confirms the exact commit exists on GitHub, refuses existing version tags, uploads all assets to a draft, and publishes after upload succeeds. GitHub creates `v<version>` at that checked commit; no manual tag push is required.

`VERSION` is a guard rather than an override. You can select another repository:

```bash
make release REPO=owner/repository
```

## Assets and recovery

A release contains four `.tar.gz` archives for macOS and Linux amd64/arm64, one Windows amd64 `.zip`, and `checksums.txt`. Versions containing `-` are marked as prereleases.

After upload failure, inspect the Releases page for incomplete drafts before retrying. If publication fails, uploaded assets remain in a draft for review. The command never automatically deletes drafts or overwrites existing tags or assets.

Example URL:

```text
https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1/erp-agent-skill_0.1.1_darwin_arm64.tar.gz
```

Published archives remain immutable. v0.1.1 contains the README, Skill, and templates from its original release commit. Later bilingual changes only reach downloaded binaries after a new version is built and published; do not replace existing v0.1.1 assets. Verify that bootstrap-script tags, archive versions, and checksum files match after releasing.

Pilot new versions with a small group before broader rollout.

## Documentation maintenance

Provide parallel Chinese and English user guides with language-switch links. Keep `README.md` and existing document paths as Chinese entry points, and use `.en.md` counterparts for English. Maintain one discoverable bilingual `SKILL.md` and bilingual references. Keep CLI help, installer messages, and WorkBuddy display fields understandable in both languages; retain command names, field identifiers, error codes, and environment-variable names unchanged.

Prefer prebuilt packages and one-command installation, followed by manual downloads and then source builds. Agent prompts should specify version selection, platform detection, checksums, safe credential entry, reconnecting, and truthful completion criteria.

Document persistent user-level authentication as the newer installer’s default. Users restart the client and run codex / claude normally or open a local desktop Code session using the same configuration. The launcher remains a compatibility option; published v0.1.1 still uses its legacy credential flow.

## Personal Keys and auditing

Issue separate Keys per user and environment, bound to the correct ERP identity, tenant, and minimum capabilities. Do not share Keys or embed them in Connector packages.

New Codex / Claude Code installations persist the Key in private user-level MCP configuration with Unix mode 0600. Never package these files, commit them, or store them in shared project configuration. The installer does not pass Keys through process arguments, and new installations need no OS keyring. Existing keyring entries are read only for compatibility. WorkBuddy users still enter Keys in its Token UI.

ERP Open Platform audits Client, Key identifier, bound user, tenant, capability, protocol, request ID, outcome, and duration. Maintainers retain checksums, tag/commit identity, rollout groups, and rollback records. Logs and support material must not contain Key secrets.

## Rollback and credential incidents

Reinstall a previous verified version and its integration. Rolling back a Skill does not change server authorization; credentials are managed by the newer private user configuration or WorkBuddy Token UI. Older releases may still need keyring credentials, so check the target version’s authentication flow before rolling back.

For suspected compromise, revoke the Key in ERP Open Platform first, inspect audit records, and issue a replacement as appropriate. Local uninstall is not server-side revocation.

## Persistent-authentication release verification

Verify replacement of legacy environment-variable registration, direct CLI startup, and local desktop Code sessions using user configuration. Cover Key rotation, uninstall, private file permissions, and preservation of unrelated MCP settings. CLI configuration-read checks do not prove actual desktop UI behavior; report those validation scopes separately.
