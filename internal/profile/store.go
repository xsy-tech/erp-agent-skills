package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store interface {
	Set(agent, endpoint string) error
	Get(agent string) (string, error)
	Delete(agent string) error
}

type FileStore struct {
	path string
	mu   sync.Mutex
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Set(agent, endpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles, err := s.load()
	if err != nil {
		return err
	}
	profiles[agent] = endpoint
	return s.save(profiles)
}

func (s *FileStore) Get(agent string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles, err := s.load()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimSpace(profiles[agent])
	if endpoint == "" {
		return "", fmt.Errorf("no configured endpoint for %s", agent)
	}
	return endpoint, nil
}

func (s *FileStore) Delete(agent string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	profiles, err := s.load()
	if err != nil {
		return err
	}
	delete(profiles, agent)
	return s.save(profiles)
}

func (s *FileStore) load() (map[string]string, error) {
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read endpoint profiles: %w", err)
	}
	profiles := make(map[string]string)
	if err := json.Unmarshal(content, &profiles); err != nil {
		return nil, fmt.Errorf("decode endpoint profiles: %w", err)
	}
	return profiles, nil
}

func (s *FileStore) save(profiles map[string]string) error {
	content, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("encode endpoint profiles: %w", err)
	}
	directory := filepath.Dir(s.path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create profile directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".profiles-")
	if err != nil {
		return fmt.Errorf("create temporary profile: %w", err)
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err := temporary.Write(append(content, '\n')); err != nil {
		temporary.Close()
		return fmt.Errorf("write endpoint profiles: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("set endpoint profile mode: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close endpoint profiles: %w", err)
	}
	if err := os.Rename(name, s.path); err != nil {
		return fmt.Errorf("replace endpoint profiles: %w", err)
	}
	return nil
}
