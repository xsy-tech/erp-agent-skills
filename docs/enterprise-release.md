# Enterprise Release

## Repository And Access

Host this repository in the enterprise Codeup organization. Limit write and tag permissions to maintainers. Ordinary users install immutable release archives and should not need repository write access.

## Release Procedure

1. Run `make check`.
2. Run `make dist`.
3. Verify every archive against `dist/checksums.txt`.
4. Review `git grep` output for credential patterns.
5. Tag the reviewed commit and publish the archives and checksum file together.
6. Pilot with a small user group before broad distribution.

Prefer signed tags and CI-generated artifacts. Consumers must pin an explicit version and verify SHA-256 before installation.

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
