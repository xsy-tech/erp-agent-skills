package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRejectsUnsafeArchivePaths(t *testing.T) {
	commandPath := filepath.Join(t.TempDir(), "distpack")
	build := exec.Command("go", "build", "-o", commandPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build distpack: %v\n%s", err, output)
	}

	tests := []struct {
		name       string
		root       string
		binaryName string
	}{
		{name: "parent root", root: "..", binaryName: "erp-skill"},
		{name: "parent binary", root: "release", binaryName: "../outside-binary"},
		{name: "windows drive", root: `C:\release`, binaryName: "erp-skill"},
		{name: "windows drive relative", root: "C:release", binaryName: "erp-skill"},
		{name: "UNC root", root: `\\server\share`, binaryName: "erp-skill"},
		{name: "nested binary", root: "release", binaryName: "nested/erp-skill"},
		{name: "control root", root: "release\nname", binaryName: "erp-skill"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			sourceDir := filepath.Join(root, "source")
			if err := os.Mkdir(sourceDir, 0o755); err != nil {
				t.Fatalf("create source directory: %v", err)
			}
			for _, name := range []string{"LICENSE", "README.md", "README.en.md", "manifest.json", "erp-skill"} {
				if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(name+"\n"), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", name, err)
				}
			}
			outsidePath := filepath.Join(root, "outside-binary")
			if err := os.WriteFile(outsidePath, []byte("outside sentinel\n"), 0o755); err != nil {
				t.Fatalf("write outside sentinel: %v", err)
			}
			outputPath := filepath.Join(root, "release.zip")

			command := exec.Command(commandPath,
				"-format", "zip",
				"-source", sourceDir,
				"-output", outputPath,
				"-root", test.root,
				"-binary", test.binaryName,
			)
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatalf("distpack accepted unsafe paths; output:\n%s", output)
			}
			if !strings.Contains(string(output), "invalid") {
				t.Fatalf("distpack output = %q, want validation error", output)
			}
			if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
				t.Fatalf("output archive exists or could not be checked: %v", statErr)
			}
		})
	}
}
