package workspaceselectflow

import "strings"

// jsTsImportInsertLine handles JavaScript / TypeScript module top matter.
func jsTsImportInsertLine(lines []string) int {
	i := 0
	for i < len(lines) {
		line := lines[i]
		t := strings.TrimSpace(line)
		if t == "" || isLeadingJsTsPreambleOrImportLine(line) {
			i++
			continue
		}
		return i
	}
	return len(lines)
}

func isLeadingJsTsPreambleOrImportLine(s string) bool {
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
	switch t {
	case `"use strict";`, `"use strict"`, `"use client";`, `"use client"`:
		return true
	}
	if strings.HasPrefix(t, "import ") && !strings.HasPrefix(t, "import(") {
		return true
	}
	if strings.HasPrefix(t, "export ") && strings.Contains(t, "from ") {
		return true
	}
	return false
}
