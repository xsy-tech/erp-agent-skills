#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"
repo=${RELEASE_REPO:-xsy-tech/erp-agent-skills}

fail() {
  printf 'Release error: %s\n' "$*" >&2
  exit 1
}

require_clean_tree() {
  [[ -z "$(git status --porcelain --untracked-files=normal)" ]] ||
    fail 'Release requires a clean working tree. Commit or move tracked and untracked changes first.'
}

for tool in git go gh; do
  command -v "$tool" >/dev/null 2>&1 || fail "Install $tool before running make release."
done
require_clean_tree
version=$(go run ./cmd/erp-skill version | awk '{print $2}')
[[ -n "$version" ]] || fail 'Cannot determine the source version.'
[[ -z "${RELEASE_VERSION:-}" || "$RELEASE_VERSION" == "$version" ]] ||
  fail "VERSION=$RELEASE_VERSION does not match source version $version. Update release metadata and commit it first."
tag="v$version"
git check-ref-format "refs/tags/$tag" >/dev/null || fail "Invalid release tag: $tag"
commit=$(git rev-parse HEAD)
gh auth status >/dev/null 2>&1 || fail 'Authenticate with gh auth login first.'
if ! remote_commit=$(gh api "repos/$repo/commits/$commit" --jq '.sha'); then
  fail "Push the current commit to $repo first, and verify your GitHub access."
fi
[[ "$remote_commit" == "$commit" ]] || fail "Push the current commit to $repo first."
if ! refs=$(gh api "repos/$repo/git/matching-refs/tags/$tag" --paginate --jq '.[].ref'); then
  fail "Cannot inspect remote tags in $repo; no release was created."
fi
while IFS= read -r ref; do
  [[ "$ref" != "refs/tags/$tag" ]] || fail "$tag already exists in $repo. Use a new source version; existing releases are never overwritten."
done <<< "$refs"

printf 'Preparing %s at %s for %s\n' "$tag" "$commit" "$repo"
"${MAKE:-make}" release-check
require_clean_tree
[[ "$(git rev-parse HEAD)" == "$commit" ]] || fail 'HEAD changed during the build. Run make release again.'

assets=()
for target in darwin_amd64 darwin_arm64 linux_amd64 linux_arm64 windows_amd64; do
  extension=tar.gz
  [[ "$target" != windows_* ]] || extension=zip
  asset="dist/erp-agent-skill_${version}_${target}.${extension}"
  [[ -f "$asset" ]] || fail "Missing release asset: $asset"
  assets+=("$asset")
done
assets+=(dist/checksums.txt)
options=()
[[ "$version" != *-* ]] || options+=(--prerelease)

# Keep incomplete uploads private. Publish only after all six assets upload.
if ! gh release create "$tag" "${assets[@]}" \
  --repo "$repo" --target "$commit" --title "ERP Agent Skill $tag" \
  --generate-notes --draft ${options[@]+"${options[@]}"}; then
  fail "Upload failed. Inspect any draft at https://github.com/$repo/releases before retrying; it was not published by this command."
fi
if ! gh release edit "$tag" --repo "$repo" --draft=false; then
  fail "Assets uploaded, but publication failed. Inspect the draft at https://github.com/$repo/releases."
fi
printf 'Published: https://github.com/%s/releases/tag/%s\n' "$repo" "$tag"
