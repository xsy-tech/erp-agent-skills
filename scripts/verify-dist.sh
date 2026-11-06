#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root/dist"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum -c checksums.txt
else
  shasum -a 256 -c checksums.txt
fi
