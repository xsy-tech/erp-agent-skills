# Enterprise Release

## Repository And Access

Releases are published to `xsy-tech/erp-agent-skills` on GitHub by default. Limit write and tag permissions to maintainers. Ordinary users install immutable release archives and should not need repository write access.

## Release Procedure

1. Update `version.go`, `manifest.json`, `scripts/install.sh`, and `scripts/install.ps1` together for a new version. `make check` checks their consistency.
2. Run `make release-check`. This runs tests and `go vet`, builds all five platform archives, and verifies `dist/checksums.txt`. It does not contact GitHub or create tags/releases.
3. Review the source and artifacts for credentials, then commit and push all release changes. Move or commit untracked files too; publishing requires a clean working tree.
4. Install GitHub CLI (`gh`) and authenticate once with `gh auth login`, using an account with write access to the repository.
5. Run `make release`, or `make release VERSION=0.1.1` to assert an expected version. The command repeats the local checks, verifies the commit exists in GitHub, uploads all assets to a draft Release, and then publishes it. GitHub creates `v<version>` pointing to that exact commit. No manual tag push is required.
6. Pilot with a small user group before broad distribution.

```bash
make help
make release-check
make release VERSION=0.1.1
# To publish a fork or another organization repository:
make release REPO=owner/repository
```

`VERSION` is a guard, not a version override; it must match the source version.
`REPO` defaults to `xsy-tech/erp-agent-skills`. The checked commit must already
exist in that repository, and the version tag must not exist. Releases never
overwrite existing tags or assets. `make release-check` can run with local edits;
`make release` checks for changes both before and after building.

Assets are four `.tar.gz` archives (`darwin/amd64`, `darwin/arm64`, `linux/amd64`,
`linux/arm64`), one `.zip` archive (`windows/amd64`), and `checksums.txt`.
Version strings containing `-` are marked as prereleases.

If uploading fails, inspect the Releases page for an incomplete draft before
retrying. If the final publication fails, the uploaded assets remain in a draft
that can be reviewed and published on GitHub. The command prints recovery
instructions and never automatically deletes drafts or existing tags.

Consumers must pin an explicit version and verify SHA-256 before installation.
For example, the Apple Silicon archive for `v0.1.1` is available after publication at
`https://github.com/xsy-tech/erp-agent-skills/releases/download/v0.1.1/erp-agent-skill_0.1.1_darwin_arm64.tar.gz`.

## API Key Provisioning

Issue one personal Key per user and intended environment. Bind it to the correct ERP user and tenant, then grant only the required Capabilities. Do not share one Key across a team or embed a Key in a Connector package.

Codex and Claude Code use the OS credential manager:

- macOS: Keychain.
- Windows: Credential Manager.
- Linux desktop: Secret Service.

Headless Linux hosts require an available Secret Service-compatible backend. Installation must fail with an actionable Keyring error when no backend is available; do not fall back to plaintext files.

WorkBuddy users enter their own Key through the Connector Token UI.

## Rollout And Rollback

Roll out by version-pinned user group. To roll back, reinstall the previous verified archive; the personal Key remains in the OS credential manager. Skill Bundle rollback does not change server-side Capability authorization.

For a credential incident, revoke the Key in ERP Open Platform first, review audit records, and issue a replacement only after the affected environment is secured. A client-side uninstall is not a substitute for server-side revocation.

## Audit Expectations

ERP Open Platform records Client, API Key, bound user, tenant, Capability, protocol, request ID, result, and timing. Release operators should retain artifact checksums, tag identity, rollout group, and rollback events. Never include the Key secret in logs or support tickets.
