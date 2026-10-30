package workbuddy

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	bundle "github.com/go-saas/erp-agent-skill"
	"github.com/go-saas/erp-agent-skill/internal/assets"
)

type templateData struct {
	Endpoint string
}

func Generate(outputDirectory, endpoint string, force bool) error {
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	generated, err := connectorFiles(endpoint)
	if err != nil {
		return err
	}
	for name, content := range generated {
		target := filepath.Join(outputDirectory, filepath.FromSlash(name))
		existing, readErr := os.ReadFile(target)
		if readErr == nil && !bytes.Equal(existing, content) && !force {
			return fmt.Errorf("refusing to overwrite changed connector file %s without --force", target)
		}
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return fmt.Errorf("inspect connector file %s: %w", target, readErr)
		}
	}
	for name, content := range generated {
		target := filepath.Join(outputDirectory, filepath.FromSlash(name))
		existing, readErr := os.ReadFile(target)
		if readErr == nil && bytes.Equal(existing, content) {
			continue
		}
		if err := writeAtomic(target, content); err != nil {
			return err
		}
	}
	return nil
}

func connectorFiles(endpoint string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, name := range []string{"connector-meta.json", "token-schema.json", "icon.svg"} {
		content, err := fs.ReadFile(bundle.ConnectorFS, "adapters/workbuddy/"+name)
		if err != nil {
			return nil, fmt.Errorf("read connector template %s: %w", name, err)
		}
		result[name] = content
	}
	mcpTemplate, err := fs.ReadFile(bundle.ConnectorFS, "adapters/workbuddy/mcp.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read MCP connector template: %w", err)
	}
	parsed, err := template.New("mcp.json").Parse(string(mcpTemplate))
	if err != nil {
		return nil, fmt.Errorf("parse MCP connector template: %w", err)
	}
	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, templateData{Endpoint: endpoint}); err != nil {
		return nil, fmt.Errorf("render MCP connector template: %w", err)
	}
	result["mcp.json"] = rendered.Bytes()
	skillFiles, err := assets.Files()
	if err != nil {
		return nil, err
	}
	for name, content := range skillFiles {
		result[filepath.ToSlash(filepath.Join("skills", "erp-open-platform", name))] = content
	}
	return result, nil
}

func validateEndpoint(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("WorkBuddy MCP endpoint must be an absolute URL")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	host := parsed.Hostname()
	if parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1") {
		return nil
	}
	return errors.New("WorkBuddy MCP endpoint must use HTTPS outside localhost")
}

func writeAtomic(target string, content []byte) error {
	directory := filepath.Dir(target)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create connector directory %s: %w", directory, err)
	}
	temporary, err := os.CreateTemp(directory, ".tmp-")
	if err != nil {
		return fmt.Errorf("create temporary connector file: %w", err)
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary connector file: %w", err)
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("set connector file mode: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary connector file: %w", err)
	}
	if err := os.Rename(name, target); err != nil {
		return fmt.Errorf("replace connector file %s: %w", target, err)
	}
	return nil
}
