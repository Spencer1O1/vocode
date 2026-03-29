package agentcontext

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"
)

func TestReadActiveFileExcerpt_emptyPath(t *testing.T) {
	t.Parallel()
	if got := ReadActiveFileExcerpt(""); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestReadActiveFileExcerpt_truncatesByRunes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "f.go")
	var b []rune
	for len(b) < maxActiveFileExcerptRunes+50 {
		b = append(b, '世')
	}
	if err := os.WriteFile(p, []byte(string(b)), 0o600); err != nil {
		t.Fatal(err)
	}
	got := ReadActiveFileExcerpt(p)
	if utf8.RuneCountInString(got) != maxActiveFileExcerptRunes {
		t.Fatalf("rune len got %d want %d", utf8.RuneCountInString(got), maxActiveFileExcerptRunes)
	}
}
