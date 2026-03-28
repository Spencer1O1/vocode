package workspace

import (
	"path/filepath"
	"testing"
)

func TestResolveTargetPath(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	abs := filepath.Join(wd, "deep", "file.go")
	rel := filepath.Join("pkg", "a.go")

	t.Run("empty target is active file", func(t *testing.T) {
		got := ResolveTargetPath("/ws", abs, "")
		if got != filepath.Clean(abs) {
			t.Fatalf("got %q want %q", got, filepath.Clean(abs))
		}
	})

	t.Run("absolute target", func(t *testing.T) {
		got := ResolveTargetPath("/ws", abs, abs)
		if got != filepath.Clean(abs) {
			t.Fatalf("got %q want %q", got, filepath.Clean(abs))
		}
	})

	t.Run("relative joins workspace", func(t *testing.T) {
		got := ResolveTargetPath(wd, abs, rel)
		want := filepath.Clean(filepath.Join(wd, rel))
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("empty workspace root with relative target", func(t *testing.T) {
		if got := ResolveTargetPath("", abs, rel); got != "" {
			t.Fatalf("got %q want empty", got)
		}
	})
}
