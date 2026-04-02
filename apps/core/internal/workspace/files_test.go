package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListWorkspaceFiles_filtersFilesAndSorts(t *testing.T) {
	root := t.TempDir()

	// Directories.
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	paths := []string{
		filepath.Join(root, "b.txt"),
		filepath.Join(sub, "a.go"),
		filepath.Join(sub, "c.js"),
	}
	for _, p := range paths {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	out, err := ListWorkspaceFiles(root, 0)
	if err != nil {
		t.Fatalf("ListWorkspaceFiles error: %v", err)
	}

	if len(out) != len(paths) {
		t.Fatalf("expected %d files, got %d", len(paths), len(out))
	}

	// Sorted lexicographically by absolute path.
	want := append([]string(nil), paths...)
	for i := range want {
		want[i] = filepath.Clean(want[i])
	}
	sort.Strings(want)
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("at index %d: expected %q, got %q", i, want[i], out[i])
		}
	}
}

