package installer

import "context"

func RegisterClaudeCode(ctx context.Context, runner CommandRunner, endpoint string) error {
	return runner.Run(ctx, "claude", "mcp", "add", "--scope", "user", "--transport", "http", "erp-open", endpoint, "--header", "Authorization: Bearer ${ERP_OPEN_CLAUDE_KEY}")
}
