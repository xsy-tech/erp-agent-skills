package credential

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	keyring "github.com/zalando/go-keyring"
)

const serviceName = "erp-agent-skill"

type backend interface {
	Set(service, account, secret string) error
	Get(service, account string) (string, error)
	Delete(service, account string) error
}

type osBackend struct{}

func (osBackend) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (osBackend) Get(service, account string) (string, error) {
	return keyring.Get(service, account)
}

func (osBackend) Delete(service, account string) error {
	return keyring.Delete(service, account)
}

type KeyringStore struct {
	backend backend
}

func NewOSStore() *KeyringStore {
	return NewKeyringStore(osBackend{})
}

func NewKeyringStore(value backend) *KeyringStore {
	return &KeyringStore{backend: value}
}

func (s *KeyringStore) Set(agent, serverURL, secret string) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("credential secret is required")
	}
	account, err := accountName(agent, serverURL)
	if err != nil {
		return err
	}
	if err := s.backend.Set(serviceName, account, secret); err != nil {
		return fmt.Errorf("store credential for %s: %w", agent, err)
	}
	return nil
}

func (s *KeyringStore) Get(agent, serverURL string) (string, error) {
	account, err := accountName(agent, serverURL)
	if err != nil {
		return "", err
	}
	secret, err := s.backend.Get(serviceName, account)
	if err != nil {
		return "", fmt.Errorf("load credential for %s: %w", agent, err)
	}
	return secret, nil
}

func (s *KeyringStore) Delete(agent, serverURL string) error {
	account, err := accountName(agent, serverURL)
	if err != nil {
		return err
	}
	if err := s.backend.Delete(serviceName, account); err != nil {
		return fmt.Errorf("delete credential for %s: %w", agent, err)
	}
	return nil
}

func accountName(agent, serverURL string) (string, error) {
	agent = strings.TrimSpace(agent)
	if agent == "" {
		return "", errors.New("credential agent is required")
	}
	normalized, err := normalizeServerURL(serverURL)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(normalized))
	return agent + ":" + hex.EncodeToString(digest[:]), nil
}

func normalizeServerURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("parse server URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("server URL must include scheme and host")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String(), nil
}
