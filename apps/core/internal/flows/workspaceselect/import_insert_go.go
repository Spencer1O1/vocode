package workspaceselectflow

import "strings"

// goImportInsertLine places new import text for Go sources.
func goImportInsertLine(lines []string) int {
	if idx, ok := goGroupedImportCloseParenLine(lines); ok {
		return idx
	}
	return goLeadingImportInsertLine(lines)
}

// goLeadingImportInsertLine scans from the top, skipping package, comments, and import blocks.
func goLeadingImportInsertLine(lines []string) int {
	i := 0
	for i < len(lines) {
		line := lines[i]
		t := strings.TrimSpace(line)

		if goImportParenBlockStart(t) {
			depth := parenDelta(line)
			for depth > 0 {
				i++
				if i >= len(lines) {
					return len(lines)
				}
				depth += parenDelta(lines[i])
			}
			i++
			continue
		}

		if t == "" || isLeadingGoPreambleOrImportLine(line) {
			i++
			continue
		}
		return i
	}
	return len(lines)
}

func goImportParenBlockStart(t string) bool {
	t = strings.TrimSpace(t)
	if !strings.HasPrefix(t, "import") {
		return false
	}
	rest := strings.TrimSpace(t[len("import"):])
	return strings.HasPrefix(rest, "(")
}

func parenDelta(s string) int {
	return strings.Count(s, "(") - strings.Count(s, ")")
}

// goGroupedImportCloseParenLine finds the first Go "import (" block and returns the line index
// of its closing ")" (insert before this line to add specs inside the block).
func goGroupedImportCloseParenLine(lines []string) (int, bool) {
	for i := 0; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if !goImportParenBlockStart(t) {
			continue
		}
		depth := parenDelta(lines[i])
		j := i
		for depth > 0 {
			j++
			if j >= len(lines) {
				return 0, false
			}
			depth += parenDelta(lines[j])
		}
		return j, true
	}
	return 0, false
}

func isLeadingGoPreambleOrImportLine(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return true
	}
	if strings.HasPrefix(t, "//") {
		return true
	}
	if strings.HasPrefix(t, "/*") || t == "*/" {
		return true
	}
	if strings.HasPrefix(t, "*") && !strings.HasPrefix(t, "*/") {
		return true
	}
	if strings.HasPrefix(t, "package ") {
		return true
	}
	if strings.HasPrefix(t, "import") && !goImportParenBlockStart(t) {
		// Single-line import (e.g. import "fmt", import . "x").
		return true
	}
	return false
}
