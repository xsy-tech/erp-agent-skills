package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBinaryInstallScript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Bash installer is for macOS and Linux")
	}
	shasum, err := exec.LookPath("shasum")
	if err != nil {
		t.Skip("fixture requires shasum")
	}
	for _, tc := range []struct {
		name, baseURL string
		sha256Only    bool
		corrupt       bool
	}{
		{name: "default GitHub release"},
		{name: "custom mirror", baseURL: "https://downloads.example.invalid/erp"},
		{name: "Linux sha256sum", baseURL: "https://downloads.example.invalid/erp", sha256Only: true},
		{name: "checksum failure preserves installation", baseURL: "https://downloads.example.invalid/erp", corrupt: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			bin, installed := filepath.Join(root, "commands"), filepath.Join(root, "installed")
			for _, path := range []string{bin, installed} {
				if err := os.MkdirAll(path, 0755); err != nil {
					t.Fatal(err)
				}
			}
			write := func(path, text string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(text), 0755); err != nil {
					t.Fatal(err)
				}
			}
			archiveName := "erp-agent-skill_" + Version + "_linux_amd64.tar.gz"
			payload := []byte("#!/bin/sh\nprintf 'erp-skill " + Version + "\\n'\n")
			var archive bytes.Buffer
			gz := gzip.NewWriter(&archive)
			tw := tar.NewWriter(gz)
			name := strings.TrimSuffix(archiveName, ".tar.gz") + "/erp-skill"
			if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(payload))}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(payload); err != nil {
				t.Fatal(err)
			}
			if err := tw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("%x", sha256.Sum256(archive.Bytes()))
			if tc.corrupt {
				digest = strings.Repeat("0", 64)
			}
			write(filepath.Join(root, archiveName), archive.String())
			write(filepath.Join(root, "checksums.txt"), digest+"  ./"+archiveName+"\n")
			write(filepath.Join(installed, "erp-skill"), "previous installation\n")
			write(filepath.Join(bin, "uname"), "#!/bin/sh\ncase \"$1\" in -s) echo Linux;; -m) echo x86_64;; esac\n")
			write(filepath.Join(bin, "curl"), `#!/usr/bin/env bash
set -eu
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o) output=$2; shift 2 ;;
    --*) shift ;;
    *) url=$1; shift ;;
  esac
done
printf '%s\n' "$url" >> "$INSTALL_TEST_ROOT/requests"
cp "$INSTALL_TEST_ROOT/${url##*/}" "$output"
`)
			// Provide sha256sum even on macOS; fail shasum to exercise Linux-only availability.
			write(filepath.Join(bin, "sha256sum"), fmt.Sprintf("#!/bin/sh\nexec %q -a 256 \"$@\"\n", shasum))
			if tc.sha256Only {
				write(filepath.Join(bin, "shasum"), "#!/bin/sh\nexit 90\n")
			}
			cmd := exec.Command("bash", "scripts/install.sh")
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "ERP_SKILL_") && !strings.HasPrefix(entry, "PATH=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "ERP_SKILL_INSTALL_DIR="+installed, "INSTALL_TEST_ROOT="+root)
			if tc.baseURL != "" {
				cmd.Env = append(cmd.Env, "ERP_SKILL_RELEASE_BASE_URL="+tc.baseURL)
			}
			output, runErr := cmd.CombinedOutput()
			content, err := os.ReadFile(filepath.Join(installed, "erp-skill"))
			if err != nil {
				t.Fatal(err)
			}
			if tc.corrupt {
				if runErr == nil || string(content) != "previous installation\n" {
					t.Fatalf("checksum failure replaced installation: %v\n%s", runErr, output)
				}
				return
			}
			if runErr != nil || !bytes.Equal(content, payload) {
				t.Fatalf("binary installation failed: %v\n%s", runErr, output)
			}
			base := tc.baseURL
			if base == "" {
				base = "https://github.com/xsy-tech/erp-agent-skills/releases/download/v" + Version
			}
			requests, err := os.ReadFile(filepath.Join(root, "requests"))
			if err != nil {
				t.Fatal(err)
			}
			if string(requests) != base+"/"+archiveName+"\n"+base+"/checksums.txt\n" {
				t.Fatalf("unexpected download URLs: %s", requests)
			}
		})
	}
}
