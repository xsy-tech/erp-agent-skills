package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var semanticVersion = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?$`)

var supportedAdapters = map[string]struct{}{
	"codex":       {},
	"claude-code": {},
	"workbuddy":   {},
}

type Manifest struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Skill    string   `json:"skill"`
	Adapters []string `json:"adapters"`
}

func Load(reader io.Reader) (Manifest, error) {
	var value Manifest
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return value, nil
}

func Validate(value Manifest) error {
	if strings.TrimSpace(value.Name) == "" {
		return errors.New("manifest name is required")
	}
	if !semanticVersion.MatchString(value.Version) {
		return errors.New("manifest version must be semantic version")
	}
	if strings.TrimSpace(value.Skill) == "" {
		return errors.New("manifest skill entry point is required")
	}
	if len(value.Adapters) == 0 {
		return errors.New("manifest must include at least one adapter")
	}
	for _, adapter := range value.Adapters {
		if _, ok := supportedAdapters[adapter]; !ok {
			return fmt.Errorf("unsupported adapter %q", adapter)
		}
	}
	return nil
}
