#!/usr/bin/env bash
set -euo pipefail

version=${ERP_SKILL_VERSION:-0.1.0}
base_url=${ERP_SKILL_RELEASE_BASE_URL:?set ERP_SKILL_RELEASE_BASE_URL to the release directory URL}
install_dir=${ERP_SKILL_INSTALL_DIR:-"$HOME/.local/bin"}

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "unsupported operating system" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  arm64|aarch64) arch=arm64 ;;
  x86_64|amd64) arch=amd64 ;;
  *) echo "unsupported architecture" >&2; exit 1 ;;
esac

archive="erp-agent-skill_${version}_${os}_${arch}.tar.gz"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
curl --fail --silent --show-error --location "$base_url/$archive" -o "$temporary/$archive"
curl --fail --silent --show-error --location "$base_url/checksums.txt" -o "$temporary/checksums.txt"
expected=$(awk -v name="$archive" '$2 == "./" name || $2 == name {print $1}' "$temporary/checksums.txt")
actual=$(shasum -a 256 "$temporary/$archive" | awk '{print $1}')
[[ -n "$expected" && "$expected" == "$actual" ]] || { echo "checksum verification failed" >&2; exit 1; }
tar -xzf "$temporary/$archive" -C "$temporary"
mkdir -p "$install_dir"
install -m 0755 "$temporary/erp-agent-skill_${version}_${os}_${arch}/erp-skill" "$install_dir/erp-skill"
echo "Installed $install_dir/erp-skill"
echo "Next: printf '%s\\n' \"\$ERP_OPEN_API_KEY\" | erp-skill install codex --url <MCP_URL> --key-stdin"
