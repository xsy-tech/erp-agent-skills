package distpack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteRejectsUnsafeArchiveNames(t *testing.T) {
	unsafeNames := []string{
		"",
		".",
		"..",
		"../outside",
		`..\outside`,
		"/absolute",
		`C:\outside`,
		"C:/outside",
		"C:outside",
		`\\server\share`,
		"//server/share",
		"nested/name",
		`nested\name`,
		"nul\x00name",
		"line\nbreak",
		"control\x1fname",
	}

	for _, field := range []string{"root", "binary"} {
		for _, name := range unsafeNames {
			t.Run(field+"/"+quotedTestName(name), func(t *testing.T) {
				sourceDir, outputPath := prepareArchiveFixture(t)
				options := Options{
					Format:     FormatZIP,
					SourceDir:  sourceDir,
					OutputPath: outputPath,
					RootName:   "release",
					BinaryName: "erp-skill",
				}
				wantError := "invalid archive root"
				if field == "root" {
					options.RootName = name
				} else {
					options.BinaryName = name
					wantError = "invalid binary name"
				}

				err := Write(options)
				if err == nil {
					t.Fatalf("Write accepted unsafe %s name %q", field, name)
				}
				if !strings.Contains(err.Error(), wantError) {
					t.Fatalf("Write error = %q, want %q", err, wantError)
				}
				assertFileDoesNotExist(t, outputPath)
			})
		}
	}
}

func TestWriteDoesNotReadBinaryOutsideSourceDir(t *testing.T) {
	sourceDir, outputPath := prepareArchiveFixture(t)
	outsidePath := filepath.Join(filepath.Dir(sourceDir), "outside-binary")
	if err := os.WriteFile(outsidePath, []byte("outside sentinel\n"), 0o755); err != nil {
		t.Fatalf("write outside sentinel: %v", err)
	}

	err := Write(Options{
		Format:     FormatZIP,
		SourceDir:  sourceDir,
		OutputPath: outputPath,
		RootName:   "release",
		BinaryName: "../outside-binary",
	})
	if err == nil {
		t.Fatal("Write accepted an outside binary source path")
	}
	if !strings.Contains(err.Error(), "invalid binary name") {
		t.Fatalf("Write error = %q, want invalid binary name", err)
	}
	assertFileDoesNotExist(t, outputPath)
}

func prepareArchiveFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	sourceDir := filepath.Join(root, "source")
	if err := os.Mkdir(sourceDir, 0o755); err != nil {
		t.Fatalf("create source directory: %v", err)
	}
	for _, name := range []string{"LICENSE", "README.md", "manifest.json", "erp-skill"} {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(name+"\n"), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	return sourceDir, filepath.Join(root, "release.zip")
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("output archive %s exists or could not be checked: %v", path, err)
	}
}

func quotedTestName(name string) string {
	return strings.NewReplacer(
		"/", "_slash_",
		`\`, "_backslash_",
		"\x00", "_nul_",
		"\n", "_newline_",
		"\x1f", "_control_",
	).Replace(name)
}
