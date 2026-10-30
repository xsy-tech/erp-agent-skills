package cli

import (
	"context"
	"errors"
	"flag"
	"io"
	"strings"

	"github.com/go-saas/erp-agent-skill/internal/installer"
)

func (a *App) runUninstall(ctx context.Context, args []string) error {
	if len(args) == 0 || (args[0] != "codex" && args[0] != "claude-code") {
		return errors.New("uninstall requires codex or claude-code")
	}
	agent := args[0]
	flags := flag.NewFlagSet("uninstall "+agent, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" {
		return errors.New("uninstall requires --url")
	}
	var err error
	if agent == "codex" {
		err = installer.UnregisterCodex(ctx, a.Commands)
	} else {
		err = installer.UnregisterClaudeCode(ctx, a.Commands)
	}
	if err != nil {
		return err
	}
	if err := installer.UninstallSkill(agent, a.HomeDir, a.CodexHome); err != nil {
		return err
	}
	if err := a.Credentials.Delete(agent, *endpoint); err != nil {
		return err
	}
	return a.Profiles.Delete(agent)
}
