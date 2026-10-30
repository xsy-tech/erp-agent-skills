package bundle

import "embed"

// ManifestJSON is the release manifest embedded in the erp-skill binary.
//
//go:embed manifest.json
var ManifestJSON []byte

// SkillFS contains the shared Skill instructions installed for supported agents.
//
//go:embed SKILL.md references/*.md
var SkillFS embed.FS
