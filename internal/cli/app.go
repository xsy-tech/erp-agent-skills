package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-saas/erp-agent-skill/internal/credential"
	"github.com/go-saas/erp-agent-skill/internal/installer"
	"github.com/go-saas/erp-agent-skill/internal/mcpclient"
	"github.com/go-saas/erp-agent-skill/internal/profile"
	secureRunner "github.com/go-saas/erp-agent-skill/internal/runner"
)

type Doctor interface {
	Doctor(context.Context, string, string) (mcpclient.DoctorResult, error)
}

type Launcher func(context.Context, string, []string, string, string) error

type App struct {
	Version     string
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	HomeDir     string
	CodexHome   string
	Credentials credential.Store
	Profiles    profile.Store
	Doctor      Doctor
	Commands    installer.CommandRunner
	Launch      Launcher
}

func NewDefault(version string) (*App, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home: %w", err)
	}
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	return &App{
		Version:     version,
		In:          os.Stdin,
		Out:         os.Stdout,
		Err:         os.Stderr,
		HomeDir:     home,
		CodexHome:   os.Getenv("CODEX_HOME"),
		Credentials: credential.NewOSStore(),
		Profiles:    profile.NewFileStore(filepath.Join(configDirectory, "erp-agent-skill", "profiles.json")),
		Doctor:      mcpclient.New(),
		Commands:    execCommandRunner{},
		Launch:      secureRunner.Run,
	}, nil
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		a.writeHelp()
		return nil
	}
	switch args[0] {
	case "version":
		fmt.Fprintf(a.Out, "erp-skill %s\n", a.Version)
		return nil
	case "install":
		return a.runInstall(ctx, args[1:])
	case "doctor":
		return a.runDoctor(ctx, args[1:])
	case "launch":
		return a.runLaunch(ctx, args[1:])
	case "uninstall":
		return a.runUninstall(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (a *App) writeHelp() {
	fmt.Fprintln(a.Out, `erp-skill manages ERP Open Platform agent integrations.

Commands:
  install codex|claude-code --url URL --key-stdin [--force]
  install workbuddy --url URL --output DIRECTORY [--force]
  doctor codex|claude-code --url URL
  doctor workbuddy --url URL --key-stdin
  launch codex|claude-code [-- agent arguments]
  uninstall codex|claude-code --url URL
  version`)
}

func readSecret(reader io.Reader) (string, error) {
	content, err := io.ReadAll(io.LimitReader(reader, 16*1024))
	if err != nil {
		return "", errors.New("read API key from stdin")
	}
	secret := strings.TrimSpace(string(content))
	if secret == "" {
		return "", errors.New("API key from stdin is empty")
	}
	if strings.ContainsRune(secret, '\x00') {
		return "", errors.New("API key from stdin is invalid")
	}
	return secret, nil
}

func redactError(err error, secret string) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if secret != "" {
		message = strings.ReplaceAll(message, secret, "[REDACTED]")
	}
	return errors.New(message)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func restoreCredential(store credential.Store, agent, endpoint, previous string, existed bool) {
	if existed {
		_ = store.Set(agent, endpoint, previous)
		return
	}
	_ = store.Delete(agent, endpoint)
}

func restoreProfile(store profile.Store, agent, previous string, existed bool) {
	if existed {
		_ = store.Set(agent, previous)
		return
	}
	_ = store.Delete(agent)
}
