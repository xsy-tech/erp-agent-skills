package credential

import (
	"errors"
	"strings"
	"testing"
)

type memoryBackend struct {
	values map[string]string
	err    error
}

func (b *memoryBackend) Set(service, account, secret string) error {
	if b.err != nil {
		return b.err
	}
	if b.values == nil {
		b.values = make(map[string]string)
	}
	b.values[service+"/"+account] = secret
	return nil
}

func (b *memoryBackend) Get(service, account string) (string, error) {
	if b.err != nil {
		return "", b.err
	}
	value, ok := b.values[service+"/"+account]
	if !ok {
		return "", errors.New("not found")
	}
	return value, nil
}

func (b *memoryBackend) Delete(service, account string) error {
	if b.err != nil {
		return b.err
	}
	delete(b.values, service+"/"+account)
	return nil
}

func TestAccountIsStableAndSeparated(t *testing.T) {
	first, err := accountName("codex", "HTTPS://ERP.EXAMPLE.COM/mcp/")
	if err != nil {
		t.Fatal(err)
	}
	second, err := accountName("codex", "https://erp.example.com/mcp")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("normalized accounts differ: %q != %q", first, second)
	}
	claude, _ := accountName("claude-code", "https://erp.example.com/mcp")
	otherServer, _ := accountName("codex", "https://other.example.com/mcp")
	if first == claude || first == otherServer {
		t.Fatal("accounts must be separated by agent and server")
	}
}

func TestStoreRejectsEmptySecretAndNeverLeaksSecret(t *testing.T) {
	backend := &memoryBackend{err: errors.New("backend unavailable")}
	store := NewKeyringStore(backend)
	if err := store.Set("codex", "https://erp.example.com/mcp", ""); err == nil {
		t.Fatal("Set() error = nil, want error")
	}
	secret := "recognizable-secret-value"
	err := store.Set("codex", "https://erp.example.com/mcp", secret)
	if err == nil {
		t.Fatal("Set() error = nil, want backend error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked secret: %v", err)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	store := NewKeyringStore(&memoryBackend{})
	const endpoint = "https://erp.example.com/mcp"
	if err := store.Set("codex", endpoint, "secret"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("codex", endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if got != "secret" {
		t.Fatalf("Get() = %q, want secret", got)
	}
	if err := store.Delete("codex", endpoint); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("codex", endpoint); err == nil {
		t.Fatal("Get() after Delete error = nil, want error")
	}
}
