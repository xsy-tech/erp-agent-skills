package bundle

import (
	"encoding/json"
	"io/fs"
	"os"
	"regexp"
	"testing"
)

func TestEmbeddedReleaseMetadataMatchesVersion(t *testing.T) {
	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(ManifestJSON, &manifest); err != nil {
		t.Fatalf("parse embedded manifest: %v", err)
	}

	connectorJSON, err := fs.ReadFile(ConnectorFS, "adapters/workbuddy/connector-meta.json")
	if err != nil {
		t.Fatalf("read embedded connector metadata: %v", err)
	}
	var connector struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(connectorJSON, &connector); err != nil {
		t.Fatalf("parse embedded connector metadata: %v", err)
	}

	for _, metadata := range []struct {
		name    string
		version string
	}{
		{name: "manifest", version: manifest.Version},
		{name: "connector", version: connector.Version},
	} {
		if metadata.version != Version {
			t.Errorf("%s version = %q, want %q", metadata.name, metadata.version, Version)
		}
	}
}

func TestInstallerDefaultVersionsMatchVersion(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		assignment string
	}{
		{
			name:       "shell",
			path:       "scripts/install.sh",
			assignment: `(?m)^[ \t]*version=\$\{ERP_SKILL_VERSION:-([^}\r\n]+)\}[ \t]*\r?$`,
		},
		{
			name:       "powershell",
			path:       "scripts/install.ps1",
			assignment: `(?m)^[ \t]*\[string\]\$Version[ \t]*=[ \t]*\$\(if[ \t]*\(\$env:ERP_SKILL_VERSION\)[ \t]*\{[ \t]*\$env:ERP_SKILL_VERSION[ \t]*\}[ \t]*else[ \t]*\{[ \t]*"([^"\r\n]+)"[ \t]*\}[ \t]*\),?[ \t]*\r?$`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatalf("read installer script: %v", err)
			}
			matches := regexp.MustCompile(test.assignment).FindAllSubmatch(content, -1)
			if len(matches) != 1 || len(matches[0]) != 2 {
				t.Fatalf("default version assignment matches = %d, want exactly 1", len(matches))
			}
			if got := string(matches[0][1]); got != Version {
				t.Errorf("default version = %q, want %q", got, Version)
			}
		})
	}
}
