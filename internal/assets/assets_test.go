package assets

import (
	"regexp"
	"strings"
	"testing"
)

func TestSkillBundleIntegrity(t *testing.T) {
	files, err := Files()
	if err != nil {
		t.Fatalf("Files() error = %v", err)
	}
	skill, ok := files["SKILL.md"]
	if !ok {
		t.Fatal("SKILL.md is missing")
	}
	if len(skill) >= 12*1024 {
		t.Fatalf("SKILL.md size = %d, want below 12288", len(skill))
	}
	if !strings.Contains(string(skill), "tools/list") {
		t.Fatal("SKILL.md must identify tools/list as the runtime authority")
	}
	for _, name := range []string{"workflows", "terminology", "safety", "errors"} {
		for _, suffix := range []string{".md", ".en.md"} {
			if _, ok := files["references/"+name+suffix]; !ok {
				t.Errorf("missing installed language resource: references/%s%s", name, suffix)
			}
		}
	}

	linkPattern := regexp.MustCompile(`\]\((references/[^)]+)\)`)
	for _, match := range linkPattern.FindAllStringSubmatch(string(skill), -1) {
		if _, ok := files[match[1]]; !ok {
			t.Errorf("linked reference %q is missing", match[1])
		}
	}

	secretPattern := regexp.MustCompile(`Authorization: Bearer\s+[A-Za-z0-9_-]{16,}`)
	for name, content := range files {
		text := string(content)
		if strings.Contains(text, "erp_local_") {
			t.Errorf("%s contains an API key prefix", name)
		}
		if secretPattern.MatchString(text) {
			t.Errorf("%s contains a literal bearer credential", name)
		}
		if strings.Contains(text, "固定 Capability 清单") {
			t.Errorf("%s contains a fixed capability inventory", name)
		}
	}
}
