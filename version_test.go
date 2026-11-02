package bundle

import (
	"encoding/json"
	"io/fs"
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
