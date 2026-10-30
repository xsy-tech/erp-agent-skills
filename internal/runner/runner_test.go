package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInjectsSecretOnlyThroughEnvironment(t *testing.T) {
	if os.Getenv("ERP_SKILL_HELPER") == "1" {
		output := os.Getenv("ERP_SKILL_OUTPUT")
		data := strings.Join(os.Args[1:], "|") + "\n" + os.Getenv("ERP_OPEN_CODEX_KEY")
		if err := os.WriteFile(output, []byte(data), 0o600); err != nil {
			os.Exit(3)
		}
		os.Exit(0)
	}
	output := filepath.Join(t.TempDir(), "output")
	t.Setenv("ERP_SKILL_HELPER", "1")
	t.Setenv("ERP_SKILL_OUTPUT", output)
	const secret = "recognizable-secret-value"
	if err := Run(context.Background(), os.Args[0], []string{"agent-arg"}, "ERP_OPEN_CODEX_KEY", secret); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(data), "\n")
	if strings.Contains(parts[0], secret) {
		t.Fatalf("secret leaked into child arguments: %q", parts[0])
	}
	if len(parts) < 2 || parts[1] != secret {
		t.Fatalf("child environment secret = %q, want supplied secret", parts[1])
	}
}
