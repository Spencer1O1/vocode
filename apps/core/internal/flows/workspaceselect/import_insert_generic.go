package workspaceselectflow

import (
	"strings"
	"unicode"
)

// genericImportInsertLine is used for unknown extensions: conservative mix of common patterns
// (no language-specific block structures like Go import ( ... )).
func genericImportInsertLine(lines []string) int {
	i := 0
	for i < len(lines) {
		line := lines[i]
		t := strings.TrimSpace(line)
		if t == "" || isLeadingGenericPreambleOrImportLine(line) {
			i++
			continue
		}
		return i
	}
	return len(lines)
}

func isLeadingGenericPreambleOrImportLine(s string) bool {
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
	if strings.HasPrefix(t, "#") {
		return true
	}
	switch t {
	case `"use strict";`, `"use strict"`, `"use client";`, `"use client"`:
		return true
	}
	if strings.HasPrefix(t, "package ") {
		return true
	}
	if strings.HasPrefix(t, "import ") && !strings.HasPrefix(t, "import(") {
		return true
	}
	if strings.HasPrefix(t, "export ") && strings.Contains(t, "from ") {
		return true
	}
	lt := strings.TrimLeftFunc(s, unicode.IsSpace)
	if strings.HasPrefix(lt, "from ") && strings.Contains(t, " import") {
		return true
	}
	return false
}
