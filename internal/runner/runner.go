package runner

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, command string, args []string, environmentKey, secret string) error {
	if strings.TrimSpace(command) == "" {
		return errors.New("agent command is required")
	}
	if strings.TrimSpace(environmentKey) == "" {
		return errors.New("credential environment variable is required")
	}
	if secret == "" {
		return errors.New("credential secret is required")
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = withEnvironment(os.Environ(), environmentKey, secret)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func withEnvironment(environment []string, key, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)
	for _, item := range environment {
		if !strings.HasPrefix(item, prefix) {
			result = append(result, item)
		}
	}
	return append(result, prefix+value)
}
