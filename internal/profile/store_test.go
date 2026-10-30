package profile

import (
	"path/filepath"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "profiles.json"))
	if err := store.Set("codex", "https://erp.example.com/mcp"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://erp.example.com/mcp" {
		t.Fatalf("Get() = %q", got)
	}
	if err := store.Delete("codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("codex"); err == nil {
		t.Fatal("Get() after Delete error = nil")
	}
}
