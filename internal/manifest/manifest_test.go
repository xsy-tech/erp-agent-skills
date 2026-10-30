package manifest

import (
	"strings"
	"testing"
)

func TestLoadAndValidate(t *testing.T) {
	input := `{"name":"erp-open-platform","version":"0.1.0","skill":"SKILL.md","adapters":["codex","claude-code","workbuddy"]}`
	got, err := Load(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := Validate(got); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidManifest(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
	}{
		{name: "empty name", manifest: Manifest{Version: "0.1.0", Skill: "SKILL.md", Adapters: []string{"codex"}}},
		{name: "invalid version", manifest: Manifest{Name: "erp", Version: "latest", Skill: "SKILL.md", Adapters: []string{"codex"}}},
		{name: "missing skill", manifest: Manifest{Name: "erp", Version: "0.1.0", Adapters: []string{"codex"}}},
		{name: "unsupported adapter", manifest: Manifest{Name: "erp", Version: "0.1.0", Skill: "SKILL.md", Adapters: []string{"unknown"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.manifest); err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
		})
	}
}
