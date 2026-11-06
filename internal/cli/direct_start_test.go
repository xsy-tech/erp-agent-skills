package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestInstallPersistsDesktopReadyMCPWithoutCommandLineSecrets(t *testing.T) {
	for _, agent := range []string{"codex", "claude-code"} {
		t.Run(agent, func(t *testing.T) {
			app, stdout, _, _, commands, _ := newTestApp(t)
			path := filepath.Join(app.HomeDir, ".claude.json")
			baseline := `{"theme":"dark","largeId":9007199254740993,"mcpServers":{"other":{"type":"http","url":"https://other.invalid"}}}`
			if agent == "codex" {
				path = filepath.Join(app.CodexHome, "config.toml")
				baseline = "# Preserve my settings\nmodel = 'existing-model'\n[mcp_servers.other]\nurl = 'https://other.invalid'\n"
			}
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(baseline), 0644); err != nil {
				t.Fatal(err)
			}
			endpoint := "https://erp.example.com/mcp"
			for _, secret := range []string{"first-personal-key", "rotated-personal-key"} {
				app.In = strings.NewReader(secret + "\n")
				if err := app.Run(t.Context(), []string{"install", agent, "--url", endpoint, "--key-stdin"}); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				var config map[string]any
				var server map[string]any
				var headerName string
				if agent == "codex" {
					if err := toml.Unmarshal(data, &config); err != nil {
						t.Fatal(err)
					}
					server = config["mcp_servers"].(map[string]any)["erp-open"].(map[string]any)
					headerName = "http_headers"
					if config["model"] != "existing-model" || !bytes.Contains(data, []byte("# Preserve my settings")) {
						t.Fatal("unrelated Codex settings or comments were lost")
					}
				} else {
					decoder := json.NewDecoder(bytes.NewReader(data))
					decoder.UseNumber()
					if err := decoder.Decode(&config); err != nil {
						t.Fatal(err)
					}
					server = config["mcpServers"].(map[string]any)["erp-open"].(map[string]any)
					headerName = "headers"
					if config["theme"] != "dark" || config["largeId"] != json.Number("9007199254740993") {
						t.Fatal("unrelated Claude settings were changed")
					}
				}
				headers, _ := server[headerName].(map[string]any)
				if server["url"] != endpoint || headers["Authorization"] != "Bearer "+secret {
					t.Fatal("persistent MCP authentication does not match installed credential")
				}
				if bytes.Contains(data, []byte("ERP_OPEN_CODEX_KEY")) || bytes.Contains(data, []byte("ERP_OPEN_CLAUDE_KEY")) {
					t.Fatal("configuration still requires a launcher environment variable")
				}
				if commands.name != "" || strings.Contains(stdout.String(), secret) || strings.Contains(app.Err.(*bytes.Buffer).String(), secret) {
					t.Fatal("installation launched a command or exposed a credential")
				}
				if info, err := os.Stat(path); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0600) {
					t.Fatal("credential-bearing configuration is not private")
				}
			}
		})
	}
}

func TestInstallClaudeUsesCustomConfigurationDirectory(t *testing.T) {
	app, _, _, _, _, _ := newTestApp(t)
	app.ClaudeConfigDir = t.TempDir()
	app.In = strings.NewReader("custom-client-key\n")
	if err := app.Run(t.Context(), []string{"install", "claude-code", "--url", "https://erp.invalid/mcp", "--key-stdin"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".claude.json", filepath.Join("skills", "erp-open-platform", "SKILL.md")} {
		if _, err := os.Stat(filepath.Join(app.ClaudeConfigDir, name)); err != nil {
			t.Fatal("custom Claude configuration and Skill were not installed together")
		}
	}
	if _, err := os.Stat(filepath.Join(app.HomeDir, ".claude.json")); !os.IsNotExist(err) {
		t.Fatal("default Claude configuration was changed despite custom directory")
	}
}

func TestFailedPersistentConfigurationRestoresProfileAndPreservesFile(t *testing.T) {
	app, stdout, credentials, profiles, _, _ := newTestApp(t)
	profiles.values = map[string]string{"codex": "https://previous.invalid/mcp"}
	credentials.values = map[string]string{"codex|https://previous.invalid/mcp": "old-cache"}
	path := filepath.Join(app.CodexHome, "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	const original = "Authorization = \"existing-key with no closing quote"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	app.In = strings.NewReader("new-key\n")
	err := app.Run(t.Context(), []string{"install", "codex", "--url", "https://new.invalid/mcp", "--key-stdin"})
	if err == nil || strings.Contains(err.Error(), "existing-key") || strings.Contains(err.Error(), "new-key") {
		t.Fatal("invalid configuration was accepted or a credential was disclosed")
	}
	data, _ := os.ReadFile(path)
	if string(data) != original || profiles.values["codex"] != "https://previous.invalid/mcp" || credentials.values["codex|https://previous.invalid/mcp"] != "old-cache" {
		t.Fatal("failed installation changed existing credentials or endpoint profile")
	}
	if stdout.Len() != 0 {
		t.Fatal("failed installation reported success")
	}
}

func TestCompatibilityLauncherPrefersPersistentConfigurationOverLegacyKeyring(t *testing.T) {
	app, _, credentials, _, _, launches := newTestApp(t)
	endpoint := "https://erp.invalid/mcp"
	app.In = strings.NewReader("new-config-key\n")
	if err := app.Run(t.Context(), []string{"install", "codex", "--url", endpoint, "--key-stdin"}); err != nil {
		t.Fatal(err)
	}
	credentials.values = map[string]string{"codex|" + endpoint: "stale-keyring-key"}
	if err := app.Run(t.Context(), []string{"launch", "codex"}); err != nil {
		t.Fatal(err)
	}
	if (*launches)[0].secret != "new-config-key" {
		t.Fatal("compatibility launcher used a stale credential")
	}
}
