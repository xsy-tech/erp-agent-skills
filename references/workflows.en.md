# ERP workflows

[中文](workflows.md) | English

## Search, then select

1. Search by name, account, document number, SKU, or another stable identifier.
2. Report no results without inventing data.
3. For one unambiguous match, display identifying fields before continuing.
4. For multiple candidates, show distinguishing fields and let the user choose.

## Read before writing

Read the current state before updating, submitting, posting, paying, shipping, or deleting. If the required read tool is unavailable, do not assume the current state.

## Create with idempotency

When a write requires `idempotency_key`, generate one stable key for the user's single intended operation. Reuse it only for an exact retry with identical input, never for different input.

## Confirm before submission

Before a high-impact call, explain the selected object, quantities or amounts, destination, resulting state, and reversibility. Invoke it after explicit confirmation for that operation.

## Refresh capabilities

After an administrator changes grants, reconnect MCP or start a new session and inspect `tools/list`. No local sync command or background polling is required.
