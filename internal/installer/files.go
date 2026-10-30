package installer

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-saas/erp-agent-skill/internal/assets"
)

func InstallSkill(agent, homeDir, codexHome string, force bool) error {
	destination, err := skillDirectory(agent, homeDir, codexHome)
	if err != nil {
		return err
	}
	files, err := assets.Files()
	if err != nil {
		return err
	}
	for name, content := range files {
		target := filepath.Join(destination, filepath.FromSlash(name))
		existing, readErr := os.ReadFile(target)
		if readErr == nil && !bytes.Equal(existing, content) && !force {
			return fmt.Errorf("refusing to overwrite changed skill file %s without --force", target)
		}
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return fmt.Errorf("inspect skill file %s: %w", target, readErr)
		}
	}
	for name, content := range files {
		target := filepath.Join(destination, filepath.FromSlash(name))
		existing, readErr := os.ReadFile(target)
		if readErr == nil && bytes.Equal(existing, content) {
			continue
		}
		if err := writeAtomic(target, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func UninstallSkill(agent, homeDir, codexHome string) error {
	destination, err := skillDirectory(agent, homeDir, codexHome)
	if err != nil {
		return err
	}
	managed, err := assets.Files()
	if err != nil {
		return err
	}
	actual := make(map[string][]byte)
	err = filepath.WalkDir(destination, func(path string, entry fs.DirEntry, walkErr error) error {
		if errors.Is(walkErr, os.ErrNotExist) {
			return fs.SkipAll
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(destination, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		actual[filepath.ToSlash(relative)] = content
		return nil
	})
	if err != nil {
		return fmt.Errorf("inspect installed skill: %w", err)
	}
	if len(actual) == 0 {
		return nil
	}
	for name, content := range actual {
		expected, ok := managed[name]
		if !ok || !bytes.Equal(content, expected) {
			return fmt.Errorf("refusing to remove skill directory containing changed or unmanaged file %s", name)
		}
	}
	return os.RemoveAll(destination)
}

func skillDirectory(agent, homeDir, codexHome string) (string, error) {
	switch agent {
	case "codex":
		if codexHome == "" {
			codexHome = filepath.Join(homeDir, ".codex")
		}
		return filepath.Join(codexHome, "skills", "erp-open-platform"), nil
	case "claude-code":
		return filepath.Join(homeDir, ".claude", "skills", "erp-open-platform"), nil
	default:
		return "", fmt.Errorf("unsupported local agent %q", agent)
	}
}

func writeAtomic(target string, content []byte, mode os.FileMode) error {
	directory := filepath.Dir(target)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create skill directory %s: %w", directory, err)
	}
	temporary, err := os.CreateTemp(directory, ".tmp-")
	if err != nil {
		return fmt.Errorf("create temporary skill file: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary skill file: %w", err)
	}
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary skill file mode: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary skill file: %w", err)
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return fmt.Errorf("replace skill file %s: %w", target, err)
	}
	return nil
}
