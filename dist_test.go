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

func TestReleaseArtifactsIgnoreGitMetadata(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("distribution script requires bash")
	}

	callerRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("get caller root: %v", err)
	}
	workspace := t.TempDir()
	sourceRoot := filepath.Join(workspace, "source-without-git")
	copySourceTree(t, callerRoot, sourceRoot)
	if _, err := os.Lstat(filepath.Join(sourceRoot, ".git")); !os.IsNotExist(err) {
		t.Fatalf("source fixture unexpectedly contains .git: %v", err)
	}
	binSentinel := filepath.Join(sourceRoot, "bin", "caller-sentinel")
	distSentinel := filepath.Join(sourceRoot, "dist", "caller-sentinel")
	writeSentinel(t, binSentinel)
	writeSentinel(t, distSentinel)
	defer assertSentinel(t, binSentinel)
	defer assertSentinel(t, distSentinel)

	firstRepo := filepath.Join(workspace, "git-context-a")
	secondRepo := filepath.Join(workspace, "git-context-b")
	copySourceTree(t, sourceRoot, firstRepo)
	copySourceTree(t, sourceRoot, secondRepo)
	firstRevision := initializeIndependentRepository(t, firstRepo, "2001-01-01T00:00:00Z", "release context a")
	secondRevision := initializeIndependentRepository(t, secondRepo, "2002-02-02T00:00:00Z", "release context b")
	if firstRevision == secondRevision {
		t.Fatalf("Git fixture revisions unexpectedly match: %s", firstRevision)
	}

	runReleaseCommand(t, firstRepo, "UTC", "make", "build")
	runReleaseCommand(t, firstRepo, "UTC", "make", "dist")
	runReleaseCommand(t, secondRepo, "Asia/Kuala_Lumpur", "make", "build")
	runReleaseCommand(t, secondRepo, "Asia/Kuala_Lumpur", "make", "dist")

	firstBinary := readTestFile(t, filepath.Join(firstRepo, "bin", "erp-skill"))
	secondBinary := readTestFile(t, filepath.Join(secondRepo, "bin", "erp-skill"))
	if !bytes.Equal(firstBinary, secondBinary) {
		t.Error("make build binary differs across Git contexts")
	}
	assertNoVCSBuildInfo(t, "Git context A make build binary", firstBinary)
	assertNoVCSBuildInfo(t, "Git context B make build binary", secondBinary)

	archiveNames := []string{
		fmt.Sprintf("erp-agent-skill_%s_darwin_amd64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_darwin_arm64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_linux_amd64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_linux_arm64.tar.gz", Version),
		fmt.Sprintf("erp-agent-skill_%s_windows_amd64.zip", Version),
	}
	for _, name := range archiveNames {
		firstArchive := readTestFile(t, filepath.Join(firstRepo, "dist", name))
		secondArchive := readTestFile(t, filepath.Join(secondRepo, "dist", name))
		firstArchiveBinary := readArchiveBinary(t, name, firstArchive)
		secondArchiveBinary := readArchiveBinary(t, name, secondArchive)
		if !bytes.Equal(firstArchiveBinary, secondArchiveBinary) {
			t.Errorf("%s binary differs across Git contexts", name)
		}
		if !bytes.Equal(firstArchive, secondArchive) {
			t.Errorf("%s differs across Git contexts", name)
		}
		assertNoVCSBuildInfo(t, "Git context A "+name+" binary", firstArchiveBinary)
		assertNoVCSBuildInfo(t, "Git context B "+name+" binary", secondArchiveBinary)
	}

	firstChecksums := readTestFile(t, filepath.Join(firstRepo, "dist", "checksums.txt"))
	secondChecksums := readTestFile(t, filepath.Join(secondRepo, "dist", "checksums.txt"))
	if !bytes.Equal(firstChecksums, secondChecksums) {
		t.Error("checksums.txt differs across Git contexts")
	}
}

func copySourceTree(t *testing.T, sourceRoot, destinationRoot string) {
	t.Helper()
	err := filepath.Walk(sourceRoot, func(sourcePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceRoot, sourcePath)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(destinationRoot, info.Mode().Perm())
		}
		firstComponent := strings.Split(relativePath, string(os.PathSeparator))[0]
		if firstComponent == ".git" || firstComponent == "bin" || firstComponent == "dist" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destinationPath := filepath.Join(destinationRoot, relativePath)
		if info.IsDir() {
			return os.MkdirAll(destinationPath, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(sourcePath)
			if err != nil {
				return err
			}
			return os.Symlink(target, destinationPath)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported source entry %s", relativePath)
		}
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		return os.WriteFile(destinationPath, content, info.Mode().Perm())
	})
	if err != nil {
		t.Fatalf("copy source tree: %v", err)
	}
}

func writeSentinel(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create sentinel directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("caller output must remain untouched\n"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
}

func assertSentinel(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("read sentinel %s: %v", path, err)
		return
	}
	if string(content) != "caller output must remain untouched\n" {
		t.Errorf("sentinel %s was modified", path)
	}
}

func initializeIndependentRepository(t *testing.T, root, commitDate, message string) string {
	t.Helper()
	runGitCommand(t, root, nil, "init", "-q")
	runGitCommand(t, root, nil, "config", "user.name", "Release Test")
	runGitCommand(t, root, nil, "config", "user.email", "release-test@example.invalid")
	runGitCommand(t, root, nil, "add", "--all")
	runGitCommand(t, root, []string{
		"GIT_AUTHOR_DATE=" + commitDate,
		"GIT_COMMITTER_DATE=" + commitDate,
	}, "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", message)
	return strings.TrimSpace(runGitCommandOutput(t, root, "rev-parse", "HEAD"))
}

func runGitCommand(t *testing.T, root string, extraEnv []string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	command.Env = append(os.Environ(), extraEnv...)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), root, err, output)
	}
}

func runGitCommandOutput(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), root, err, output)
	}
	return string(output)
}

func runReleaseCommand(t *testing.T, root, timezone, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = root
	command.Env = append(os.Environ(), "LC_ALL=C", "LANG=C", "TZ="+timezone)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run %s %s in %s: %v\n%s", name, strings.Join(args, " "), root, err, output)
	}
}

func readArchiveBinary(t *testing.T, archiveName string, content []byte) []byte {
	t.Helper()
	binarySuffix := "/erp-skill"
	if strings.HasSuffix(archiveName, ".zip") {
		binarySuffix += ".exe"
		reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
		if err != nil {
			t.Fatalf("open %s: %v", archiveName, err)
		}
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, binarySuffix) {
				entry, err := file.Open()
				if err != nil {
					t.Fatalf("open %s binary: %v", archiveName, err)
				}
				binary, readErr := io.ReadAll(entry)
				closeErr := entry.Close()
				if readErr != nil {
					t.Fatalf("read %s binary: %v", archiveName, readErr)
				}
				if closeErr != nil {
					t.Fatalf("close %s binary: %v", archiveName, closeErr)
				}
				return binary
			}
		}
		t.Fatalf("find binary in %s", archiveName)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("open %s gzip stream: %v", archiveName, err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read %s tar header: %v", archiveName, err)
		}
		if strings.HasSuffix(header.Name, binarySuffix) {
			binary, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatalf("read %s binary: %v", archiveName, err)
			}
			return binary
		}
	}
	t.Fatalf("find binary in %s", archiveName)
	return nil
}

func assertNoVCSBuildInfo(t *testing.T, label string, binary []byte) {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), "erp-skill")
	if err := os.WriteFile(binaryPath, binary, 0o755); err != nil {
		t.Fatalf("write %s for build info inspection: %v", label, err)
	}
	command := exec.Command("go", "version", "-m", binaryPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("inspect %s with go version -m: %v\n%s", label, err, output)
	}
	for _, field := range []string{"vcs.revision", "vcs.time", "vcs.modified"} {
		if bytes.Contains(output, []byte(field)) {
			t.Errorf("%s contains environment-dependent %s metadata:\n%s", label, field, output)
		}
	}
}

func prepareDistFixture(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
		t.Fatalf("create fixture root: %v", err)
	}
	for _, name := range []string{"README.md", "README.en.md", "LICENSE", "manifest.json"} {
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
		{name: packageName + "/README.en.md", mode: 0o644, content: readTestFile(t, "README.en.md")},
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
