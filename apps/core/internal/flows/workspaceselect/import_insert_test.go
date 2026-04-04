package workspaceselectflow

import "testing"

func TestJsTsImportInsertLine_afterImports(t *testing.T) {
	lines := []string{
		`import { a } from "x"`,
		`import b from "y"`,
		``,
		`export function f() {}`,
	}
	if got := jsTsImportInsertLine(lines); got != 2 {
		t.Fatalf("got %d want 2 (insert after last import, at the blank line)", got)
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
	if got := jsTsImportInsertLine(lines); got != 3 {
		t.Fatalf("got %d want 3 (insert at blank line after import, before const)", got)
	}
}

func TestJsTsImportInsertLine_expoAppStyle(t *testing.T) {
	lines := []string{
		`import { StatusBar } from 'expo-status-bar';`,
		`import { StyleSheet, Text, View } from 'react-native';`,
		``,
		`export default function App() {}`,
	}
	if got := jsTsImportInsertLine(lines); got != 2 {
		t.Fatalf("got %d want 2", got)
	}
}

func TestJstsFinalizeImportBlock_insertsBlankBeforeCode(t *testing.T) {
	src := []string{
		`import { a } from "x"`,
		`export default function App() {}`,
	}
	insertLine := 1
	block := "import { useState } from 'react';\n"
	got := jstsFinalizeImportBlock(src, insertLine, block)
	want := "import { useState } from 'react';\n\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestJstsFinalizeImportBlock_skipWhenNextIsImport(t *testing.T) {
	src := []string{`import { a } from "x"`, `import { b } from "y"`}
	block := "import { c } from \"z\";\n"
	got := jstsFinalizeImportBlock(src, 1, block)
	if got != block {
		t.Fatalf("got %q want unchanged %q", got, block)
	}
}

func TestJstsFinalizeImportBlock_existingBlankBeforeExport(t *testing.T) {
	src := []string{
		`import { a } from "x"`,
		`import React from 'react';`,
		``,
		`export default function Test() {}`,
	}
	insertLine := 2
	block := "import { Button } from 'react-native';\n"
	got := jstsFinalizeImportBlock(src, insertLine, block)
	if got != block {
		t.Fatalf("got %q want %q (single file blank already separates imports from export)", got, block)
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
