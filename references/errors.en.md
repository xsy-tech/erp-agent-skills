# ERP error handling

[中文](errors.md) | English

| Error or symptom | Agent response |
| --- | --- |
| Tool absent from tools/list | Explain that the connection does not expose the capability. Do not invent a call; reconnect if authorization just changed. |
| CAPABILITY_DENIED | Explain the missing capability and ask for administrator authorization. |
| Invalid, expired, or revoked Key | Stop ERP calls and ask the user to repair their credential using the approved installer or client Token UI. |
| Tenant or user mismatch | Stop; do not override identity parameters. Ask an administrator for the correct personal Key. |
| Rate limiting | Follow retry guidance, avoid parallel retries, and explain when another attempt is appropriate. |
| Upstream timeout or outage | Report an unconfirmed outcome. Read current state before retrying writes. |
| Validation error | Correct only invalid business input; do not weaken or bypass server validation. |

Error explanations must not contain credentials, raw upstream response bodies, SQL, or stack traces.
