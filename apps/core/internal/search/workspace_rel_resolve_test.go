package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveWorkspaceRelativePath_caseAndSeparators(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Evade"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "my_utils", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, ok := ResolveWorkspaceRelativePath(root, "evade")
	if !ok {
		t.Fatal("expected evade -> Evade")
	}
	want := filepath.Join(root, "Evade")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %s want %s", got, want)
	}

	got, ok = ResolveWorkspaceRelativePath(root, "my utils/nested")
	if !ok {
		t.Fatal("expected my utils/nested")
	}
	want = filepath.Join(root, "my_utils", "nested")
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}
