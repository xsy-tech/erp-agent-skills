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
	if len(args) == 0 || (args[0] != "codex" && args[0] != "claude-code" && args[0] != "workbuddy") {
		return errors.New("uninstall requires codex, claude-code, or workbuddy")
	}
	agent := args[0]
	flags := flag.NewFlagSet("uninstall "+agent, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" {
		return errors.New("uninstall requires --url")
	}
	if err := a.clientConfig().Unregister(agent, *endpoint); err != nil {
		return err
	}
	if err := installer.UninstallSkill(agent, a.HomeDir, a.CodexHome, a.ClaudeConfigDir); err != nil {
		return err
	}
	// New installs use private client configuration. Remove a legacy keyring
	// entry when it is available, without making new installs depend on keyrings.
	if agent != "workbuddy" {
		if _, err := a.Credentials.Get(agent, *endpoint); err == nil {
			if err := a.Credentials.Delete(agent, *endpoint); err != nil {
				return err
			}
		}
	}
	return a.Profiles.Delete(agent)
}
