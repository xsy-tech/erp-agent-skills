package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDistArchivesAreCanonicalAcrossCheckoutRoots(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("distribution script requires bash")
	}

	workspace := t.TempDir()
	fakeBin := filepath.Join(workspace, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("create fake bin directory: %v", err)
	}
	fakeGo := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  run)
    printf 'erp-skill %s\n'
    ;;
  build)
    shift
    output=
    while (($#)); do
      if [[ "$1" == "-o" ]]; then
        output=$2
        shift 2
      else
        shift
      fi
    done
    [[ -n "$output" ]]
    printf 'deterministic test binary\n' > "$output"
    chmod 0755 "$output"
    ;;
  *)
    echo "unsupported fake go command: $*" >&2
    exit 1
    ;;
esac
`, Version)
	if err := os.WriteFile(filepath.Join(fakeBin, "go"), []byte(fakeGo), 0o755); err != nil {
		t.Fatalf("write fake go command: %v", err)
	}

	checkoutRoots := []string{
		filepath.Join(workspace, "checkout-a"),
		filepath.Join(workspace, "nested", "checkout-b"),
	}
	for _, root := range checkoutRoots {
		prepareDistFixture(t, root)
		command := exec.Command("bash", "scripts/dist.sh")
		command.Dir = root
		command.Env = append(os.Environ(),
			"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
			"LC_ALL=C",
			"LANG=C",
			"TZ=UTC",
		)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build distribution in %s: %v\n%s", root, err, output)
		}
	}

	archives := []string{
		fmt.Sprintf("erp-agent-skill_%s_darwin_amd64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_darwin_arm64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_linux_amd64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_linux_arm64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_windows_amd64.zip", Version),
		"checksums.txt",
	}
	for _, name := range archives {
		first := readTestFile(t, filepath.Join(checkoutRoots[0], "dist", name))
		second := readTestFile(t, filepath.Join(checkoutRoots[1], "dist", name))
		if !bytes.Equal(first, second) {
			t.Errorf("%s differs across checkout roots", name)
		}
		if strings.HasSuffix(name, ".tar.gz") {
			assertCanonicalTarHeaders(t, name, first)
		}
	}
}

func prepareDistFixture(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
		t.Fatalf("create fixture root: %v", err)
	}
	for _, name := range []string{"README.md", "LICENSE", "manifest.json"} {
		content := readTestFile(t, name)
		if err := os.WriteFile(filepath.Join(root, name), content, 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	script := readTestFile(t, "scripts/dist.sh")
	if err := os.WriteFile(filepath.Join(root, "scripts", "dist.sh"), script, 0o755); err != nil {
		t.Fatalf("write fixture distribution script: %v", err)
	}
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}

func assertCanonicalTarHeaders(t *testing.T, archiveName string, content []byte) {
	t.Helper()
	gzipReader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("open %s gzip stream: %v", archiveName, err)
	}
	defer gzipReader.Close()

	packageName := strings.TrimSuffix(archiveName, ".tar.gz")
	wantNames := []string{
		packageName + "/LICENSE",
		packageName + "/README.md",
		packageName + "/manifest.json",
		packageName + "/erp-skill",
	}
	tarReader := tar.NewReader(gzipReader)
	for index, wantName := range wantNames {
		header, err := tarReader.Next()
		if err != nil {
			t.Fatalf("read %s header %d: %v", archiveName, index, err)
		}
		if header.Name != wantName {
			t.Errorf("%s header %d name = %q, want %q", archiveName, index, header.Name, wantName)
		}
		if header.Uid != 0 || header.Gid != 0 || header.Uname != "root" || header.Gname != "root" {
			t.Errorf("%s header %q owner = %d/%d %q/%q, want 0/0 root/root", archiveName, header.Name, header.Uid, header.Gid, header.Uname, header.Gname)
		}
	}
	if _, err := tarReader.Next(); err != io.EOF {
		t.Fatalf("%s has unexpected extra tar entry or error: %v", archiveName, err)
	}
}
