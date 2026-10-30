# ERP Workflows

## Search Then Select

1. Search with the user's name, account, document number, SKU, or other stable identifier.
2. If there are no results, report that fact without inventing data.
3. If there is one unambiguous result, display the identifying fields before continuing.
4. If there are multiple candidates, display the fields that distinguish them and ask the user to select one.

## Read Before Write

Read the current object state before updating, submitting, posting, paying, shipping, or deleting it. Refuse to assume the current status when the required read tool is unavailable.

## Create With Idempotency

When a write tool requires `idempotency_key`, create one stable key for the user's single intended operation and reuse it only for an exact retry of the same input. Never reuse it for different input.

## Confirm Before Submit

Before a high-impact tool call, summarize the selected object, quantities or amounts, destination, resulting state, and whether the action can be reversed. Call the tool only after explicit user confirmation.

## Capability Refresh

Do not run a local synchronization command. If an administrator changes the API Key's capabilities, reconnect the MCP server or start a new agent session and inspect `tools/list` again.
