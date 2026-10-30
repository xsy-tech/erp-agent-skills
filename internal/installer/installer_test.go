package installer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type recordedCommand struct {
	name string
	args []string
}

type fakeRunner struct {
	commands []recordedCommand
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	r.commands = append(r.commands, recordedCommand{name: name, args: append([]string(nil), args...)})
	return nil
}

func TestInstallSkillIsIdempotentAndProtectsChangedFiles(t *testing.T) {
	home := t.TempDir()
	if err := InstallSkill("codex", home, "", false); err != nil {
		t.Fatal(err)
	}
	if err := InstallSkill("codex", home, "", false); err != nil {
		t.Fatalf("idempotent InstallSkill() error = %v", err)
	}
	destination := filepath.Join(home, ".codex", "skills", "erp-open-platform", "SKILL.md")
	if err := os.WriteFile(destination, []byte("locally changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := InstallSkill("codex", home, "", false); err == nil {
		t.Fatal("InstallSkill() error = nil for changed file")
	}
	if err := InstallSkill("codex", home, "", true); err != nil {
		t.Fatalf("forced InstallSkill() error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(destination))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Fatalf("temporary file remained after install: %s", entry.Name())
		}
	}
}

func TestInstallSkillUsesAgentSpecificDirectories(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(home, "custom-codex")
	if err := InstallSkill("codex", home, codexHome, false); err != nil {
		t.Fatal(err)
	}
	if err := InstallSkill("claude-code", home, codexHome, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(codexHome, "skills", "erp-open-platform", "SKILL.md"),
		filepath.Join(home, ".claude", "skills", "erp-open-platform", "SKILL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("installed file %s: %v", path, err)
		}
	}
}

func TestRegisterMCPCommandsContainOnlyEnvironmentPlaceholders(t *testing.T) {
	runner := &fakeRunner{}
	endpoint := "https://erp.example.com/mcp"
	if err := RegisterCodex(t.Context(), runner, endpoint); err != nil {
		t.Fatal(err)
	}
	if err := RegisterClaudeCode(t.Context(), runner, endpoint); err != nil {
		t.Fatal(err)
	}
	want := []recordedCommand{
		{name: "codex", args: []string{"mcp", "add", "erp-open", "--url", endpoint, "--bearer-token-env-var", "ERP_OPEN_CODEX_KEY"}},
		{name: "claude", args: []string{"mcp", "add", "--scope", "user", "--transport", "http", "erp-open", endpoint, "--header", "Authorization: Bearer ${ERP_OPEN_CLAUDE_KEY}"}},
	}
	if !reflect.DeepEqual(runner.commands, want) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, want)
	}
}
