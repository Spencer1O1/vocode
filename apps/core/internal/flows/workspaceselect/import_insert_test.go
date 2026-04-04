package workspaceselectflow

import "testing"

func TestJsTsImportInsertLine_afterImports(t *testing.T) {
	lines := []string{
		`import { a } from "x"`,
		`import b from "y"`,
		``,
		`export function f() {}`,
	}
	if got := jsTsImportInsertLine(lines); got != 3 {
		t.Fatalf("got %d want 3", got)
	}
}

func TestJsTsImportInsertLine_useClient(t *testing.T) {
	lines := []string{
		`"use client";`,
		``,
		`import { x } from "z"`,
		``,
		`const k = 1`,
	}
	if got := jsTsImportInsertLine(lines); got != 4 {
		t.Fatalf("got %d want 4", got)
	}
}

func TestImportInsertLine_goGroupedBeforeCloseParen(t *testing.T) {
	lines := []string{
		"package main",
		"",
		"import (",
		`	"fmt"`,
		")",
		"",
		"func main() {}",
	}
	if got := importInsertLine(lines, "/tmp/main.go"); got != 4 {
		t.Fatalf("got %d want 4 (closing paren line)", got)
	}
}

func TestImportInsertLine_goSingleImportBeforeFunc(t *testing.T) {
	lines := []string{
		"package main",
		"",
		`import "fmt"`,
		"",
		"func main() {}",
	}
	if got := importInsertLine(lines, "main.go"); got != 4 {
		t.Fatalf("got %d want 4 (before func)", got)
	}
}

func TestImportInsertLine_unknownExtUsesGeneric(t *testing.T) {
	lines := []string{`import x from "y"`, `foo()`}
	if got := importInsertLine(lines, "component.vue"); got != 1 {
		t.Fatalf("generic: got %d want 1 (after import)", got)
	}
}

func TestPythonImportInsertLine(t *testing.T) {
	lines := []string{
		"# -*- coding: utf-8 -*-",
		"",
		"import os",
		"",
		"x = 1",
	}
	if got := pythonImportInsertLine(lines); got != 4 {
		t.Fatalf("got %d want 4", got)
	}
	if importInsertLine(lines, "mod.py") != 4 {
		t.Fatalf("router should use python for .py")
	}
}

func TestImportLanguageForPath(t *testing.T) {
	tests := []struct {
		path string
		want importLang
	}{
		{"x.go", importLangGo},
		{"X.GO", importLangGo},
		{"a.ts", importLangJSTS},
		{"a.tsx", importLangJSTS},
		{"a.mjs", importLangJSTS},
		{"lib.pyi", importLangPython},
		{"README.md", importLangUnknown},
	}
	for _, tc := range tests {
		if got := importLanguageForPath(tc.path); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.path, got, tc.want)
		}
	}
}

func TestShiftRangeAfterImportInsert(t *testing.T) {
	sl, sc, el, ec := shiftRangeAfterImportInsert(10, 0, 12, 5, 0, 2)
	if sl != 12 || el != 14 || sc != 0 || ec != 5 {
		t.Fatalf("got %d,%d-%d,%d", sl, sc, el, ec)
	}
	sl, sc, el, ec = shiftRangeAfterImportInsert(1, 0, 3, 0, 5, 2)
	if sl != 1 || el != 3 {
		t.Fatalf("selection before insert should not shift: got %d-%d", sl, el)
	}
}

func TestFilterNewImportLines_dedup(t *testing.T) {
	body := `import { a } from "x"
const z = 1`
	lines := []string{`import { a } from "x"`, `import { b } from "y"`}
	got := filterNewImportLines(body, lines)
	if len(got) != 1 || got[0] != `import { b } from "y"` {
		t.Fatalf("got %#v", got)
	}
}
