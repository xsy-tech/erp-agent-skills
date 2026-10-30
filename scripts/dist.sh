#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
version=$(cd "$root" && go run ./cmd/erp-skill version | awk '{print $2}')
dist="$root/dist"
stage="$dist/.stage"

rm -rf "$dist"
mkdir -p "$stage"

targets=(
  darwin/arm64
  darwin/amd64
  linux/arm64
  linux/amd64
  windows/amd64
)

for target in "${targets[@]}"; do
  os=${target%/*}
  arch=${target#*/}
  name="erp-agent-skill_${version}_${os}_${arch}"
  package_dir="$stage/$name"
  mkdir -p "$package_dir"
  binary="$package_dir/erp-skill"
  if [[ "$os" == "windows" ]]; then
    binary+=".exe"
  fi
  (
    cd "$root"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -o "$binary" ./cmd/erp-skill
  )
  cp "$root/README.md" "$root/LICENSE" "$root/manifest.json" "$package_dir/"
  find "$package_dir" -exec touch -t 198001010000 {} +
  if [[ "$os" == "windows" ]]; then
    (
      cd "$stage"
      zip -X -q "$dist/$name.zip" \
        "$name/LICENSE" \
        "$name/README.md" \
        "$name/manifest.json" \
        "$name/erp-skill.exe"
    )
  else
    (
      cd "$stage"
      COPYFILE_DISABLE=1 tar --format ustar -cf - \
        "$name/LICENSE" \
        "$name/README.md" \
        "$name/manifest.json" \
        "$name/erp-skill" \
        | gzip -n > "$dist/$name.tar.gz"
    )
  fi
done

rm -rf "$stage"
(
  cd "$dist"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum ./*.tar.gz ./*.zip > checksums.txt
  else
    shasum -a 256 ./*.tar.gz ./*.zip > checksums.txt
  fi
)
