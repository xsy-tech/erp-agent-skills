package workbuddy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCreatesTokenConnectorWithoutLiteralSecret(t *testing.T) {
	output := filepath.Join(t.TempDir(), "connector")
	endpoint := "https://erp.example.com/mcp"
	if err := Generate(output, endpoint, false); err != nil {
		t.Fatal(err)
	}

	mcpBytes, err := os.ReadFile(filepath.Join(output, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var mcpConfig struct {
		Servers map[string]struct {
			Type    string            `json:"type"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(mcpBytes, &mcpConfig); err != nil {
		t.Fatalf("mcp.json is invalid: %v", err)
	}
	server := mcpConfig.Servers["erp-open-platform"]
	if server.Type != "streamableHttp" || server.URL != endpoint {
		t.Fatalf("unexpected MCP server: %+v", server)
	}
	if server.Headers["Authorization"] != "Bearer ${ERP_OPEN_API_KEY}" {
		t.Fatalf("Authorization = %q", server.Headers["Authorization"])
	}

	tokenBytes, err := os.ReadFile(filepath.Join(output, "token-schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var tokenSchema struct {
		Fields []struct {
			Key      string `json:"key"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(tokenBytes, &tokenSchema); err != nil {
		t.Fatalf("token-schema.json is invalid: %v", err)
	}
	if len(tokenSchema.Fields) != 1 || tokenSchema.Fields[0].Key != "ERP_OPEN_API_KEY" || tokenSchema.Fields[0].Type != "password" || !tokenSchema.Fields[0].Required {
		t.Fatalf("unexpected token schema: %+v", tokenSchema.Fields)
	}

	for _, required := range []string{
		"connector-meta.json", "mcp.json", "token-schema.json", "icon.svg",
		filepath.Join("skills", "erp-open-platform", "SKILL.md"),
	} {
		if _, err := os.Stat(filepath.Join(output, required)); err != nil {
			t.Fatalf("missing generated file %s: %v", required, err)
		}
	}
	if strings.Contains(string(mcpBytes), "erp_local_") {
		t.Fatal("generated connector contains a literal API key")
	}
}

func TestGenerateProtectsExistingConnector(t *testing.T) {
	output := filepath.Join(t.TempDir(), "connector")
	if err := Generate(output, "https://erp.example.com/mcp", false); err != nil {
		t.Fatal(err)
	}
	meta := filepath.Join(output, "connector-meta.json")
	if err := os.WriteFile(meta, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Generate(output, "https://erp.example.com/mcp", false); err == nil {
		t.Fatal("Generate() error = nil for changed connector")
	}
	if err := Generate(output, "https://erp.example.com/mcp", true); err != nil {
		t.Fatalf("forced Generate() error = %v", err)
	}
}
