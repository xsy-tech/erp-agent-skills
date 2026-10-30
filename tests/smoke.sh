#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ERP_OPEN_TEST_ENDPOINT:-}" || -z "${ERP_OPEN_TEST_KEY:-}" ]]; then
  echo "SKIP: set ERP_OPEN_TEST_ENDPOINT and ERP_OPEN_TEST_KEY for live MCP smoke testing"
  exit 0
fi

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
binary=${ERP_SKILL_BIN:-"$root/bin/erp-skill"}
if [[ ! -x "$binary" ]]; then
  echo "erp-skill binary not found at $binary" >&2
  exit 1
fi

output=$(printf '%s\n' "$ERP_OPEN_TEST_KEY" | "$binary" doctor workbuddy --url "$ERP_OPEN_TEST_ENDPOINT" --key-stdin)
if [[ "$output" == *"$ERP_OPEN_TEST_KEY"* ]]; then
  echo "doctor output leaked the API key" >&2
  exit 1
fi
printf '%s' "$output" | grep -q '"ServerName": "erp-open-platform"'
printf '%s' "$output" | grep -q '"Tools": \['
echo "PASS: MCP initialize and tools/list succeeded without leaking the key"
