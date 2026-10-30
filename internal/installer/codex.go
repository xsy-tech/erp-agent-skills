package installer

import "context"

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

func RegisterCodex(ctx context.Context, runner CommandRunner, endpoint string) error {
	return runner.Run(ctx, "codex", "mcp", "add", "erp-open", "--url", endpoint, "--bearer-token-env-var", "ERP_OPEN_CODEX_KEY")
}

func UnregisterCodex(ctx context.Context, runner CommandRunner) error {
	return runner.Run(ctx, "codex", "mcp", "remove", "erp-open")
}
