package tags

import (
	"path/filepath"
	"testing"
)

func TestParseTreeSitterCLITagLine(t *testing.T) {
	t.Parallel()
	file := filepath.Clean("/tmp/x.go")
	line := "Resolver  \t | function\tdef (4, 5) - (10, 1) `func Resolver() {}`"
	tg, ok := parseTreeSitterCLITagLine(line, file)
	if !ok {
		t.Fatal("expected modern parse")
	}
	if tg.Name != "Resolver" {
		t.Fatalf("name %q", tg.Name)
	}
	if tg.Path != filepath.Clean(file) {
		t.Fatalf("path %q", tg.Path)
	}
	if tg.Kind != "function" {
		t.Fatalf("kind %q", tg.Kind)
	}
	if !tg.IsDefinition {
		t.Fatal("want def")
	}
	if tg.StartLine != 4 || tg.StartCharacter != 5 || tg.EndLine != 10 || tg.EndCharacter != 1 {
		t.Fatalf("span %+v", tg)
	}
	if !tg.Contains(7, 0) {
		t.Fatal("expected cursor inside span")
	}
	if tg.Contains(3, 0) {
		t.Fatal("before span")
	}
}

func TestParseTreeSitterCLITagLine_refSkippedBySelect(t *testing.T) {
	t.Parallel()
	line := "foo       \t | variable\tref (1, 0) - (1, 3) `foo`"
	tg, ok := parseTreeSitterCLITagLine(line, "/x.go")
	if !ok {
		t.Fatal("parse")
	}
	if tg.IsDefinition {
		t.Fatal("expected ref")
	}
	_, okSel := SelectInnermostTag([]Tag{tg}, 1, 1)
	if okSel {
		t.Fatal("refs must not be cursor targets")
	}
}

func TestNormalizeKind(t *testing.T) {
	t.Parallel()
	if got := NormalizeKind("f"); got != "function" {
		t.Fatalf("expected function, got %q", got)
	}
	if got := NormalizeKind("Class"); got != "class" {
		t.Fatalf("expected class, got %q", got)
	}
	if got := NormalizeKind("kind:method"); got != "method" {
		t.Fatalf("expected method, got %q", got)
	}
	if got := NormalizeKind("trait"); got != "interface" {
		t.Fatalf("expected interface, got %q", got)
	}
}
