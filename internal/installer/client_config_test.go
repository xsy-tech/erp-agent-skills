package installer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPersistentClientConfigLifecycle(t *testing.T) {
	for _, agent := range []string{"codex", "claude-code", "workbuddy"} {
		t.Run(agent, func(t *testing.T) {
			config := ClientConfig{HomeDir: t.TempDir()}
			path, _ := config.Path(agent)
			initial := `{"unknown":{"keep":true},"mcpServers":{"other":{"url":"https://other.invalid"},"erp-open":{"type":"http","url":"https://erp.invalid/mcp","headers":{"Authorization":"Bearer ${ERP_OPEN_CLAUDE_KEY}"}}}}`
			if agent == "codex" {
				initial = "# user comment\nmodel = 'existing'\n[mcp_servers.other]\nurl = 'https://other.invalid'\n[mcp_servers.erp-open]\nurl = 'https://erp.invalid/mcp'\nbearer_token_env_var = 'ERP_OPEN_CODEX_KEY'\n"
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
				t.Fatal(err)
			}
			endpoint := "https://erp.invalid/mcp"
			for _, secret := range []string{`key-with-"quotes"-and-\slashes`, "rotated-credential"} {
				if err := config.Register(agent, endpoint, secret); err != nil {
					t.Fatal(err)
				}
				got, err := config.Credential(agent, endpoint)
				if err != nil || got != secret {
					t.Fatal("credential did not survive serialization")
				}
				content, _ := os.ReadFile(path)
				if bytes.Contains(content, []byte("ERP_OPEN_")) {
					t.Fatal("legacy credential placeholder was not replaced")
				}
			}
			if err := config.Unregister(agent, "https://different.invalid/mcp"); err == nil {
				t.Fatal("uninstall accepted a different endpoint")
			}
			if err := config.Unregister(agent, endpoint); err != nil {
				t.Fatal(err)
			}
			if _, err := config.Credential(agent, endpoint); !errors.Is(err, ErrConfigCredentialNotFound) {
				t.Fatal("credential remained after uninstall")
			}
			content, _ := os.ReadFile(path)
			if bytes.Contains(content, []byte("rotated-credential")) || !bytes.Contains(content, []byte("https://other.invalid")) {
				t.Fatal("uninstall did not preserve only unrelated MCP settings")
			}
		})
	}
}

func TestPersistentClientConfigRejectsInvalidFilesWithoutDisclosure(t *testing.T) {
	for _, agent := range []string{"codex", "claude-code", "workbuddy"} {
		t.Run(agent, func(t *testing.T) {
			config := ClientConfig{HomeDir: t.TempDir()}
			path, _ := config.Path(agent)
			_ = os.MkdirAll(filepath.Dir(path), 0755)
			const invalid = "invalid config with very-private-existing-key [ {"
			_ = os.WriteFile(path, []byte(invalid), 0600)
			for _, action := range []func() error{
				func() error { return config.Register(agent, "https://erp.invalid/mcp", "new-private-key") },
				func() error { return config.Unregister(agent, "https://erp.invalid/mcp") },
			} {
				err := action()
				if err == nil || strings.Contains(err.Error(), "private") {
					t.Fatal("malformed config was accepted or disclosed")
				}
				content, _ := os.ReadFile(path)
				if string(content) != invalid {
					t.Fatal("invalid configuration was overwritten")
				}
			}
		})
	}
}

func TestPersistentCodexInlineConfigurationAndMultilineValues(t *testing.T) {
	for _, initial := range []string{
		`mcp_servers = { other = { url = "https://other.invalid" }, erp-open = {url="https://old.invalid"} }`,
		"instructions = '''\n[mcp_servers.erp-open]\nThis is user text, not a table.\n'''\n[mcp_servers.other]\nurl='https://other.invalid'\n",
		"# keep\n[mcp_servers]\nother = {url='https://other.invalid'}\nerp-open = {url='https://old.invalid'}\n",
	} {
		config := ClientConfig{HomeDir: t.TempDir()}
		path, _ := config.Path("codex")
		_ = os.MkdirAll(filepath.Dir(path), 0755)
		_ = os.WriteFile(path, []byte(initial), 0600)
		if err := config.Register("codex", "https://erp.invalid/mcp", "test-key"); err != nil {
			t.Fatal(err)
		}
		content, _ := os.ReadFile(path)
		_, servers, err := decodeClientConfig("codex", content)
		if err != nil || servers["other"].(map[string]any)["url"] != "https://other.invalid" {
			t.Fatal("inline or multiline configuration was corrupted")
		}
	}
}

func TestPersistentClientConfigPathsAndSymlinkProtection(t *testing.T) {
	config := ClientConfig{HomeDir: t.TempDir(), CodexHome: t.TempDir(), ClaudeConfigDir: t.TempDir()}
	for agent, want := range map[string]string{
		"codex":       filepath.Join(config.CodexHome, "config.toml"),
		"claude-code": filepath.Join(config.ClaudeConfigDir, ".claude.json"),
	} {
		path, err := config.Path(agent)
		if err != nil || path != want {
			t.Fatal("custom client configuration path not respected")
		}
		if err := config.Register(agent, "https://erp.invalid/mcp", "test-key"); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.GOOS == "windows" {
		return
	}
	config = ClientConfig{HomeDir: t.TempDir()}
	target := filepath.Join(config.HomeDir, "shared.json")
	_ = os.WriteFile(target, []byte(`{"shared":true}`), 0644)
	path, _ := config.Path("claude-code")
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if err := config.Register("claude-code", "https://erp.invalid/mcp", "test-key"); err == nil {
		t.Fatal("credential written through a symlink")
	}
	content, _ := os.ReadFile(target)
	if string(content) != `{"shared":true}` {
		t.Fatal("shared symlink target changed")
	}
}
