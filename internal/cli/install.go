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
	if agent == "workbuddy" {
		return a.installWorkBuddy(args[1:])
	}
	if agent != "codex" && agent != "claude-code" {
		return fmt.Errorf("unsupported agent %q", agent)
	}
	flags := flag.NewFlagSet("install "+agent, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	keyStdin := flags.Bool("key-stdin", false, "read API key from stdin")
	force := flags.Bool("force", false, "overwrite managed files")
	if err := flags.Parse(args[1:]); err != nil {
		return errors.New("invalid install arguments")
	}
	if flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" {
		return errors.New("install requires --url")
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
	previousSecret, previousSecretErr := a.Credentials.Get(agent, *endpoint)
	credentialExisted := previousSecretErr == nil
	previousEndpoint, previousEndpointErr := a.Profiles.Get(agent)
	profileExisted := previousEndpointErr == nil
	if err := a.Credentials.Set(agent, *endpoint, secret); err != nil {
		return redactError(err, secret)
	}
	rollback := func() {
		restoreCredential(a.Credentials, agent, *endpoint, previousSecret, credentialExisted)
		restoreProfile(a.Profiles, agent, previousEndpoint, profileExisted)
	}
	if err := a.Profiles.Set(agent, *endpoint); err != nil {
		rollback()
		return redactError(err, secret)
	}
	if err := installer.InstallSkill(agent, a.HomeDir, a.CodexHome, *force); err != nil {
		rollback()
		return redactError(err, secret)
	}
	if agent == "codex" {
		err = installer.RegisterCodex(ctx, a.Commands, *endpoint)
	} else {
		err = installer.RegisterClaudeCode(ctx, a.Commands, *endpoint)
	}
	if err != nil {
		rollback()
		return redactError(err, secret)
	}
	fmt.Fprintf(a.Out, "installed %s integration for %s\n", agent, *endpoint)
	return nil
}

func (a *App) installWorkBuddy(args []string) error {
	flags := flag.NewFlagSet("install workbuddy", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	output := flags.String("output", "", "connector output directory")
	force := flags.Bool("force", false, "overwrite managed files")
	if err := flags.Parse(args); err != nil {
		return errors.New("invalid WorkBuddy install arguments")
	}
	if flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" || strings.TrimSpace(*output) == "" {
		return errors.New("WorkBuddy install requires --url and --output")
	}
	if err := workbuddy.Generate(*output, *endpoint, *force); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "generated WorkBuddy connector at %s\n", *output)
	return nil
}
