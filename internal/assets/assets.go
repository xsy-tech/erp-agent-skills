package assets

import (
	"fmt"
	"io/fs"

	bundle "github.com/go-saas/erp-agent-skill"
)

func Files() (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(bundle.SkillFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || (path != "SKILL.md" && path[:min(len(path), len("references/"))] != "references/") {
			return nil
		}
		content, err := fs.ReadFile(bundle.SkillFS, path)
		if err != nil {
			return fmt.Errorf("read embedded asset %s: %w", path, err)
		}
		files[path] = content
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk embedded skill assets: %w", err)
	}
	return files, nil
}
