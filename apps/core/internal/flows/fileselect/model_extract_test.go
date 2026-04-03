package fileselectflow

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestWorkspaceDirectoryHints(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "evade", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := workspaceDirectoryHints(root)
	if !slices.Contains(h, "evade") {
		t.Fatalf("missing evade, got %v", h)
	}
	nested := filepath.ToSlash(filepath.Join("evade", "nested"))
	if !slices.Contains(h, nested) {
		t.Fatalf("missing %q, got %v", nested, h)
	}
	if !slices.Contains(h, "other") {
		t.Fatalf("missing other, got %v", h)
	}
}
