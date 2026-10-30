# ERP Error Handling

| Error | Agent response |
|---|---|
| Tool absent from `tools/list` | Explain that the current connection does not expose the capability. Do not invent the call. Reconnect if authorization was just changed. |
| `CAPABILITY_DENIED` | Explain that the current API Key lacks the capability and request administrator authorization. |
| Invalid, expired, or revoked key | Stop ERP calls and ask the user to repair their personal credential through the approved installer or agent Token UI. |
| Tenant or user mismatch | Stop. Do not attempt to override identity parameters. Ask an administrator to issue the correct personal key. |
| Rate limited | Respect retry guidance, avoid parallel retries, and tell the user when another attempt is appropriate. |
| Upstream timeout or unavailable | Report the operation as unconfirmed. For writes, read the current state before any retry. |
| Validation error | Correct only the invalid business input. Do not weaken or bypass server validation. |

Never include credentials, raw upstream bodies, SQL, or stack traces when explaining an error.
