package bundle

import (
	"archive/tar"
	"archive/zip"
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
	"time"
)

var canonicalArchiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

const fakeBinaryContent = "deterministic test binary\n"

func TestDistArchivesAreCanonicalAcrossCheckoutRoots(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("distribution script requires bash")
	}

	workspace := t.TempDir()
	fakeBin := filepath.Join(workspace, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("create fake bin directory: %v", err)
	}
	distpackBinary := filepath.Join(workspace, "distpack")
	command := exec.Command("go", "build", "-o", distpackBinary, "./cmd/distpack")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build distpack command: %v\n%s", err, output)
	}
	fakeGo := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
case "${1:-}" in
  run)
    case "${2:-}" in
      ./cmd/erp-skill)
        printf 'erp-skill %s\n'
        ;;
      ./cmd/distpack)
        shift 2
        [[ -x "${DISTPACK_TEST_BINARY:-}" ]]
        exec "$DISTPACK_TEST_BINARY" "$@"
        ;;
      *)
        echo "unsupported fake go run command: $*" >&2
        exit 1
        ;;
    esac
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
		filepath.Join(workspace, "checkout-utc"),
		filepath.Join(workspace, "nested", "checkout-kuala-lumpur"),
	}
	timezones := []string{"UTC", "Asia/Kuala_Lumpur"}
	for index, root := range checkoutRoots {
		prepareDistFixture(t, root)
		command := exec.Command("bash", "scripts/dist.sh")
		command.Dir = root
		command.Env = append(os.Environ(),
			"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
			"DISTPACK_TEST_BINARY="+distpackBinary,
			"LC_ALL=C",
			"LANG=C",
			"TZ="+timezones[index],
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
			assertCanonicalTarArchive(t, name, first, expectedArchiveEntries(t, name))
		}
		if strings.HasSuffix(name, ".zip") {
			assertCanonicalZIPArchive(t, name, first, expectedArchiveEntries(t, name))
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

type expectedArchiveEntry struct {
	name    string
	mode    os.FileMode
	content []byte
}

func expectedArchiveEntries(t *testing.T, archiveName string) []expectedArchiveEntry {
	t.Helper()
	packageName := strings.TrimSuffix(strings.TrimSuffix(archiveName, ".tar.gz"), ".zip")
	binaryName := "erp-skill"
	if strings.HasSuffix(archiveName, ".zip") {
		binaryName += ".exe"
	}
	return []expectedArchiveEntry{
		{name: packageName + "/LICENSE", mode: 0o644, content: readTestFile(t, "LICENSE")},
		{name: packageName + "/README.md", mode: 0o644, content: readTestFile(t, "README.md")},
		{name: packageName + "/manifest.json", mode: 0o644, content: readTestFile(t, "manifest.json")},
		{name: packageName + "/" + binaryName, mode: 0o755, content: []byte(fakeBinaryContent)},
	}
}

func assertCanonicalTarArchive(t *testing.T, archiveName string, content []byte, entries []expectedArchiveEntry) {
	t.Helper()
	gzipReader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("open %s gzip stream: %v", archiveName, err)
	}
	defer gzipReader.Close()
	if !gzipReader.Header.ModTime.IsZero() || gzipReader.Header.Name != "" || gzipReader.Header.Comment != "" || len(gzipReader.Header.Extra) != 0 || gzipReader.Header.OS != 255 {
		t.Errorf("%s gzip header is not canonical: %+v", archiveName, gzipReader.Header)
	}

	tarReader := tar.NewReader(gzipReader)
	for index, want := range entries {
		header, err := tarReader.Next()
		if err != nil {
			t.Fatalf("read %s header %d: %v", archiveName, index, err)
		}
		if header.Name != want.name {
			t.Errorf("%s header %d name = %q, want %q", archiveName, index, header.Name, want.name)
		}
		if header.Uid != 0 || header.Gid != 0 || header.Uname != "root" || header.Gname != "root" {
			t.Errorf("%s header %q owner = %d/%d %q/%q, want 0/0 root/root", archiveName, header.Name, header.Uid, header.Gid, header.Uname, header.Gname)
		}
		if header.Mode != int64(want.mode) {
			t.Errorf("%s header %q mode = %#o, want %#o", archiveName, header.Name, header.Mode, want.mode)
		}
		if header.Typeflag != tar.TypeReg {
			t.Errorf("%s header %q type = %d, want regular file", archiveName, header.Name, header.Typeflag)
		}
		if !header.ModTime.Equal(canonicalArchiveTime) {
			t.Errorf("%s header %q time = %s, want %s", archiveName, header.Name, header.ModTime, canonicalArchiveTime)
		}
		payload, err := io.ReadAll(tarReader)
		if err != nil {
			t.Fatalf("read %s payload %q: %v", archiveName, header.Name, err)
		}
		if !bytes.Equal(payload, want.content) {
			t.Errorf("%s payload %q differs", archiveName, header.Name)
		}
	}
	if _, err := tarReader.Next(); err != io.EOF {
		t.Fatalf("%s has unexpected extra tar entry or error: %v", archiveName, err)
	}
}

func assertCanonicalZIPArchive(t *testing.T, archiveName string, content []byte, entries []expectedArchiveEntry) {
	t.Helper()
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("open %s: %v", archiveName, err)
	}
	if len(zipReader.File) != len(entries) {
		t.Fatalf("%s entries = %d, want %d", archiveName, len(zipReader.File), len(entries))
	}
	for index, want := range entries {
		file := zipReader.File[index]
		if file.Name != want.name {
			t.Errorf("%s header %d name = %q, want %q", archiveName, index, file.Name, want.name)
		}
		if file.Mode().Perm() != want.mode {
			t.Errorf("%s header %q mode = %#o, want %#o", archiveName, file.Name, file.Mode().Perm(), want.mode)
		}
		if file.Method != zip.Deflate {
			t.Errorf("%s header %q method = %d, want deflate", archiveName, file.Name, file.Method)
		}
		if !file.Modified.Equal(canonicalArchiveTime) {
			t.Errorf("%s header %q time = %s, want %s", archiveName, file.Name, file.Modified, canonicalArchiveTime)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatalf("open %s payload %q: %v", archiveName, file.Name, err)
		}
		payload, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			t.Fatalf("read %s payload %q: %v", archiveName, file.Name, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close %s payload %q: %v", archiveName, file.Name, closeErr)
		}
		if !bytes.Equal(payload, want.content) {
			t.Errorf("%s payload %q differs", archiveName, file.Name)
		}
	}
}
