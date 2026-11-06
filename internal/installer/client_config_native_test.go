package installer

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in checks run actual local clients against isolated configuration files.
// They do not invoke inference or touch the user's real credential stores.
func TestNativeClientsReadPersistentMCPWithoutLauncher(t *testing.T) {
	if os.Getenv("ERP_TEST_NATIVE_CLIENTS") != "1" {
		t.Skip("set ERP_TEST_NATIVE_CLIENTS=1 to check installed client CLIs")
	}
	for _, agent := range []string{"codex", "claude-code"} {
		t.Run(agent, func(t *testing.T) {
			command := "codex"
			if agent == "claude-code" {
				command = "claude"
			}
			binary, err := exec.LookPath(command)
			if err != nil {
				t.Skip("client not installed")
			}
			home := t.TempDir()
			config := ClientConfig{HomeDir: home, CodexHome: filepath.Join(home, "codex"), ClaudeConfigDir: filepath.Join(home, "claude")}
			const endpoint = "https://erp-desktop-test.invalid/mcp"
			const secret = "isolated-native-client-test-key"
			if err := config.Register(agent, endpoint, secret); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			args := []string{"mcp", "get", "erp-open"}
			if agent == "codex" {
				args = append(args, "--json")
			}
			cmd := exec.CommandContext(ctx, binary, args...)
			cmd.Dir = home
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "CODEX_HOME=") && !strings.HasPrefix(value, "CLAUDE_CONFIG_DIR=") && !strings.HasPrefix(value, "ERP_OPEN_") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			cmd.Env = append(cmd.Env, "CODEX_HOME="+config.CodexHome, "CLAUDE_CONFIG_DIR="+config.ClaudeConfigDir)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native client could not read isolated MCP configuration: %v", err)
			}
			if agent == "codex" {
				var parsed struct {
					Transport struct {
						URL               string            `json:"url"`
						BearerTokenEnvVar *string           `json:"bearer_token_env_var"`
						HTTPHeaders       map[string]string `json:"http_headers"`
					} `json:"transport"`
				}
				if err := json.Unmarshal(output, &parsed); err != nil {
					t.Fatal("native Codex returned unexpected MCP metadata")
				}
				if parsed.Transport.URL != endpoint || parsed.Transport.BearerTokenEnvVar != nil || parsed.Transport.HTTPHeaders["Authorization"] != "Bearer "+secret {
					t.Fatal("native Codex did not load the fixed authentication header")
				}
			} else if !strings.Contains(string(output), endpoint) || !strings.Contains(string(output), "Authorization") || strings.Contains(string(output), "ERP_OPEN_CLAUDE_KEY") {
				t.Fatal("native Claude Code did not load the saved HTTP server and authentication header")
			}
		})
	}
}
