package bundle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMakeRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release tooling requires Bash and Make")
	}
	for _, tc := range []struct {
		name      string
		env       []string
		args      []string
		dirty     string
		wantError string
		wantDraft bool
	}{
		{name: "publish"},
		{name: "version mismatch", args: []string{"VERSION=9.9.9"}, wantError: "does not match"},
		{name: "tracked changes", dirty: "README.md", wantError: "clean working tree"},
		{name: "untracked changes", dirty: "untracked.go", wantError: "clean working tree"},
		{name: "unpublished commit", env: []string{"UNPUBLISHED=1"}, wantError: "Push the current commit"},
		{name: "existing tag", env: []string{"EXISTING_TAG=1"}, wantError: "already exists"},
		{name: "tag lookup failure", env: []string{"TAG_LOOKUP_FAIL=1"}, wantError: "Cannot inspect remote tags"},
		{name: "checks fail", env: []string{"CHECK_FAIL=1"}, wantError: "Error"},
		{name: "source changes during build", env: []string{"BUILD_DIRTY=1"}, wantError: "clean working tree"},
		{name: "upload fails", env: []string{"UPLOAD_FAIL=1"}, wantError: "draft", wantDraft: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, env, logPath := releaseFixture(t)
			if tc.dirty != "" {
				if err := os.WriteFile(filepath.Join(root, tc.dirty), []byte("changed"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("make", append([]string{"release"}, tc.args...)...)
			cmd.Dir = root
			cmd.Env = append(env, tc.env...)
			output, err := cmd.CombinedOutput()
			if tc.wantError == "" && err != nil {
				t.Fatalf("release failed: %v\n%s", err, output)
			}
			if tc.wantError != "" && (err == nil || !strings.Contains(string(output), tc.wantError)) {
				t.Fatalf("expected failure containing %q, got %v\n%s", tc.wantError, err, output)
			}
			log, _ := os.ReadFile(logPath)
			created := strings.Contains(string(log), "release create")
			published := strings.Contains(string(log), "--draft=false")
			if tc.wantError != "" {
				if created != tc.wantDraft || published {
					t.Fatalf("unexpected external write:\n%s", log)
				}
				return
			}
			if !created || !published {
				t.Fatalf("expected draft creation followed by publication:\n%s", log)
			}
			for _, asset := range releaseAssetNames() {
				if !strings.Contains(string(log), "dist/"+asset) {
					t.Errorf("missing release asset %s:\n%s", asset, log)
				}
			}
			sha := strings.TrimSpace(runGitCommandOutput(t, root, "rev-parse", "HEAD"))
			if !strings.Contains(string(log), "--target "+sha) || !strings.Contains(string(log), "--draft") {
				t.Fatalf("release must target the checked commit and start as a draft:\n%s", log)
			}
		})
	}
}

func TestMakeReleaseCheckDoesNotPublish(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release tooling requires Bash and Make")
	}
	root, env, logPath := releaseFixture(t)
	cmd := exec.Command("make", "-j4", "release-check")
	cmd.Dir, cmd.Env = root, env
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("release-check failed: %v\n%s", err, output)
	}
	if log, _ := os.ReadFile(logPath); len(log) != 0 {
		t.Fatalf("release-check invoked GitHub:\n%s", log)
	}
	for _, name := range releaseAssetNames() {
		if _, err := os.Stat(filepath.Join(root, "dist", name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMakeReleaseDryRunDoesNotPublish(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release tooling requires Bash and Make")
	}
	root, env, logPath := releaseFixture(t)
	cmd := exec.Command("make", "-n", "release")
	cmd.Dir, cmd.Env = root, env
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dry run failed: %v\n%s", err, output)
	}
	if log, _ := os.ReadFile(logPath); len(log) != 0 {
		t.Fatalf("make -n release invoked GitHub:\n%s", log)
	}
}

func releaseAssetNames() []string {
	var names []string
	for _, target := range []string{"darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64", "windows_amd64"} {
		ext := ".tar.gz"
		if strings.HasPrefix(target, "windows") {
			ext = ".zip"
		}
		names = append(names, "erp-agent-skill_"+Version+"_"+target+ext)
	}
	return append(names, "checksums.txt")
}

func releaseFixture(t *testing.T) (string, []string, string) {
	t.Helper()
	workspace := t.TempDir()
	root, bin := filepath.Join(workspace, "repo"), filepath.Join(workspace, "bin")
	for _, dir := range []string{filepath.Join(root, "scripts"), bin} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Makefile", "scripts/release.sh", "scripts/verify-dist.sh"} {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(root, name), string(content))
	}
	write(filepath.Join(root, ".gitignore"), "/dist/\n")
	write(filepath.Join(root, "README.md"), "release fixture\n")
	write(filepath.Join(bin, "go"), fmt.Sprintf(`#!/usr/bin/env bash
set -eu
case "$1" in
  run) printf 'erp-skill %s\n' ;;
  test) test "${CHECK_FAIL:-0}" != 1 ;;
  vet) ;;
  *) exit 97 ;;
esac
`, Version))
	write(filepath.Join(bin, "gh"), `#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >> "$RELEASE_TEST_LOG"
case "$1 $2" in
  'auth status') ;;
  'api repos/'*)
    case "$2" in
      */commits/*) test "${UNPUBLISHED:-0}" != 1; git rev-parse HEAD ;;
      */git/matching-refs/tags/*)
        test "${TAG_LOOKUP_FAIL:-0}" != 1
        if [[ "${EXISTING_TAG:-0}" == 1 ]]; then printf 'refs/tags/%s\n' "${2##*/}"; fi ;;
      *) exit 96 ;;
    esac ;;
  'release create') test "${UPLOAD_FAIL:-0}" != 1 ;;
  'release edit') ;;
  *) exit 95 ;;
esac
`)
	var dist strings.Builder
	dist.WriteString("#!/usr/bin/env bash\nset -eu\nmkdir -p dist\n")
	for _, name := range releaseAssetNames()[:5] {
		fmt.Fprintf(&dist, "printf 'fixture' > dist/%s\n", name)
	}
	dist.WriteString("cd dist\nif command -v sha256sum >/dev/null; then sha256sum ./*.tar.gz ./*.zip; else shasum -a 256 ./*.tar.gz ./*.zip; fi > checksums.txt\n")
	dist.WriteString("if [[ ${BUILD_DIRTY:-0} == 1 ]]; then printf changed >> ../README.md; fi\n")
	write(filepath.Join(root, "scripts/dist.sh"), dist.String())
	runGitCommand(t, root, nil, "init", "--quiet")
	runGitCommand(t, root, nil, "add", ".")
	runGitCommand(t, root, nil, "-c", "user.name=Release Test", "-c", "user.email=release@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
	logPath := filepath.Join(workspace, "gh.log")
	env := append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "RELEASE_TEST_LOG="+logPath)
	return root, env, logPath
}
