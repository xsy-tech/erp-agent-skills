package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/go-saas/erp-agent-skill/internal/installer"
	"github.com/go-saas/erp-agent-skill/internal/workbuddy"
)

func (a *App) runInstall(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("install requires an agent")
	}
	agent := args[0]
	if agent != "codex" && agent != "claude-code" && agent != "workbuddy" {
		return fmt.Errorf("unsupported agent %q", agent)
	}
	flags := flag.NewFlagSet("install "+agent, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	keyStdin := flags.Bool("key-stdin", false, "read API key from stdin")
	force := flags.Bool("force", false, "overwrite managed files")
	var output string
	if agent == "workbuddy" {
		flags.StringVar(&output, "output", "", "export an optional Connector directory instead of installing locally")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return errors.New("invalid install arguments")
	}
	if flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" {
		return errors.New("install requires --url")
	}
	if output != "" {
		if *keyStdin {
			return errors.New("Connector export does not accept --key-stdin; omit --output to install WorkBuddy directly")
		}
		if err := workbuddy.Generate(output, *endpoint, *force); err != nil {
			return err
		}
		fmt.Fprintf(a.Out, "已导出可选 Connector / Exported optional Connector: %s\n", output)
		return nil
	}
	if !*keyStdin {
		return errors.New("install requires --key-stdin; literal key arguments are not supported")
	}
	secret, err := readSecret(a.In)
	if err != nil {
		return err
	}
	if _, err := a.Doctor.Doctor(ctx, *endpoint, secret); err != nil {
		return redactError(err, secret)
	}
	previousEndpoint, previousEndpointErr := a.Profiles.Get(agent)
	profileExisted := previousEndpointErr == nil
	rollback := func() {
		restoreProfile(a.Profiles, agent, previousEndpoint, profileExisted)
	}
	if err := a.Profiles.Set(agent, *endpoint); err != nil {
		rollback()
		return redactError(err, secret)
	}
	if err := installer.InstallSkill(agent, a.HomeDir, a.CodexHome, *force, a.ClaudeConfigDir); err != nil {
		rollback()
		return redactError(err, secret)
	}
	err = a.clientConfig().Register(agent, *endpoint, secret)
	if err != nil {
		rollback()
		return redactError(err, secret)
	}
	fmt.Fprintf(a.Out, "installed %s integration for %s\n", agent, *endpoint)
	fmt.Fprintln(a.Out, "重启客户端后可直接使用 / Restart your client, then use Codex, Claude Code, or WorkBuddy normally.")
	return nil
}
