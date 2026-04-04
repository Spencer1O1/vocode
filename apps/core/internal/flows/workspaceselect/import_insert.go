package workspaceselectflow

import (
	"path/filepath"
	"strings"
)

// importLang selects which top-of-file / import heuristics apply.
type importLang int

const (
	importLangUnknown importLang = iota
	importLangGo
	importLangJSTS
	importLangPython
)

func importLanguageForPath(path string) importLang {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(strings.TrimSpace(path)), "."))
	switch ext {
	case "go":
		return importLangGo
	case "py", "pyi", "pyw":
		return importLangPython
	case "js", "jsx", "mjs", "cjs", "ts", "tsx", "mts", "cts":
		return importLangJSTS
	default:
		return importLangUnknown
	}
}

// importInsertLine returns the 0-based line index for a zero-width insert (new import lines).
func importInsertLine(lines []string, filePath string) int {
	switch importLanguageForPath(filePath) {
	case importLangGo:
		return goImportInsertLine(lines)
	case importLangPython:
		return pythonImportInsertLine(lines)
	case importLangJSTS:
		return jsTsImportInsertLine(lines)
	default:
		return genericImportInsertLine(lines)
	}
}

// filterNewImportLines drops empty lines and lines whose normalized form already appears in body.
func filterNewImportLines(body string, lines []string) []string {
	norm := func(s string) string {
		return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	}
	existing := norm(body)
	out := make([]string, 0, len(lines))
	for _, L := range lines {
		t := strings.TrimSpace(L)
		if t == "" {
			continue
		}
		n := norm(t)
		if n == "" {
			continue
		}
		if strings.Contains(existing, n) {
			continue
		}
		out = append(out, t)
		existing += "\n" + n
	}
	return out
}

// importBlockForInsert joins non-empty import lines with a trailing newline (empty if none).
func importBlockForInsert(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// linesAddedByImportBlock counts how many newline-terminated lines an import block adds.
func linesAddedByImportBlock(block string) int {
	if block == "" {
		return 0
	}
	return strings.Count(block, "\n")
}

// shiftRangeAfterImportInsert adds lineOffset to start/end lines when the insert happened before them.
func shiftRangeAfterImportInsert(sl, sc, el, ec, insertLine, lineOffset int) (int, int, int, int) {
	if lineOffset <= 0 {
		return sl, sc, el, ec
	}
	if sl >= insertLine {
		sl += lineOffset
	}
	if el >= insertLine {
		el += lineOffset
	}
	return sl, sc, el, ec
}
