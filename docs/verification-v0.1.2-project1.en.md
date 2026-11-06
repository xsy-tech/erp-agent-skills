# v0.1.2 project1 verification record

[中文](verification-v0.1.2-project1.md) | English

Test date: 2026-11-06. **Overall result: partial.** Native Codex and WorkBuddy model sessions queried real ERP data. Claude Code model requests were blocked by HTTP 403, and one upstream health check failed.

## Scope

- Used the macOS arm64 archive from the [official v0.1.2 Release](https://github.com/xsy-tech/erp-agent-skills/releases/tag/v0.1.2), without rebuilding the test binary.
- Archive SHA-256: `1a7c9e62e2b618d824757eb57e08c014c1a83ce5336a9d08f4def4ea8041ea95`.
- Clients: Codex CLI 0.166.0, Claude Code 2.1.316, and WorkBuddy desktop 5.3.11.
- The local platform ran with project1 configuration, and authenticated `get_platform_info` returned `environment=project1`. Real services were used without business mocks.
- The same persistent test Key was installed in all three user-scoped MCP configurations. Its grants cover only platform information, service status, and user search. It expires on 2027-11-06; the initial short-lived test Key was revoked.
- Other operating systems, CPU architectures, and Codex/Claude desktop interfaces were not tested.

## Results

| Check | Codex | Claude Code | WorkBuddy desktop |
| --- | --- | --- | --- |
| Doctor with the saved Key | VERIFIED | VERIFIED | VERIFIED |
| Native MCP connection and Skill discovery | PASS | PASS | PASS after first-use trust |
| Model calls platform info and confirms project1 | PASS | NOT RUN: model 403 | PASS |
| Model calls user search | PASS | NOT RUN: model 403 | PASS |
| api-user health | Independent MCP control PASS | NOT RUN | PASS |
| saas-erp-backend health | FAIL: upstream unavailable | NOT RUN | FAIL: upstream unavailable |

The native Codex model session invoked platform information, user search, and service status. The native WorkBuddy desktop session invoked platform information, user search, and service status for two targets. Four server audit events corroborated three successes and one upstream failure. No business write tools were called.

Claude Code initialization confirmed the connected ERP MCP and discovered Skill. Its model request then returned `403 Request not allowed`, preventing verification of model-initiated business calls. This block is recorded separately from ERP Key authentication.

The configured saas-erp-backend target was not listening during verification, so its health check failed. Business code was not changed, and healthy responses were not fabricated.

## WorkBuddy first connection

After direct MCP and Skill installation, WorkBuddy may still require trust for the MCP service. The initial session discovered the Skill but could not discover ERP tools. Completing first-use trust in **MCP service management** enabled **3/3 tools**; a fresh native session then completed the queries.

No Connector package import or `erp-skill launch` was required. Doctor verifies the saved MCP configuration, authentication, and a read-only call. Client trust and model-account availability also require verification in the native client.

## Retained configuration and evidence

The persistent Key, user-scoped MCP configuration, Skills, and WorkBuddy trust remain available for subsequent testing. Configuration permissions were verified as `0600`, and other existing MCP settings were preserved. Temporary credential copies were removed, and this run's administrative sessions were logged out.

Only this sanitized summary is published in the repository. Raw logs, screenshots, server audit records, and identity metadata remain local and are excluded from the commit.
