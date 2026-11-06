#!/usr/bin/env bash
set -euo pipefail

version=${ERP_SKILL_VERSION:-0.1.1}
base_url=${ERP_SKILL_RELEASE_BASE_URL:-"https://github.com/xsy-tech/erp-agent-skills/releases/download/v$version"}
install_dir=${ERP_SKILL_INSTALL_DIR:-"$HOME/.local/bin"}

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) echo "不支持当前操作系统；Windows 请使用 install.ps1。" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  arm64|aarch64) arch=arm64 ;;
  x86_64|amd64) arch=amd64 ;;
  *) echo "不支持当前处理器架构；仅提供 amd64 和 arm64 安装包。" >&2; exit 1 ;;
esac

archive="erp-agent-skill_${version}_${os}_${arch}.tar.gz"
temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
curl --fail --silent --show-error --location "$base_url/$archive" -o "$temporary/$archive"
curl --fail --silent --show-error --location "$base_url/checksums.txt" -o "$temporary/checksums.txt"
expected=$(awk -v name="$archive" '$2 == "./" name || $2 == name {print $1}' "$temporary/checksums.txt")
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$temporary/$archive" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$temporary/$archive" | awk '{print $1}')
fi
[[ -n "$expected" && "$expected" == "$actual" ]] || { echo "SHA-256 校验失败，已停止安装。" >&2; exit 1; }
tar -xzf "$temporary/$archive" -C "$temporary"
mkdir -p "$install_dir"
install -m 0755 "$temporary/erp-agent-skill_${version}_${os}_${arch}/erp-skill" "$install_dir/erp-skill"
echo "已安装：$install_dir/erp-skill"
printf "当前终端加入 PATH：export PATH=\"%s:\$PATH\"\n" "$install_dir"
echo "下一步：运行 erp-skill version，再按安装指南配置 Codex、Claude Code 或 WorkBuddy。"
