package workspaceselectflow

import (
	"strings"
	"unicode"
)

// pythonImportInsertLine handles Python module top (imports and common preamble).
func pythonImportInsertLine(lines []string) int {
	i := 0
	for i < len(lines) {
		line := lines[i]
		t := strings.TrimSpace(line)
		if t == "" || isLeadingPythonPreambleOrImportLine(line) {
			i++
			continue
		}
		return i
	}
	return len(lines)
}

func isLeadingPythonPreambleOrImportLine(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return true
	}
	if strings.HasPrefix(t, "#") {
		return true
	}
	lt := strings.TrimLeftFunc(s, unicode.IsSpace)
	if strings.HasPrefix(lt, "from ") && strings.Contains(t, " import") {
		return true
	}
	if strings.HasPrefix(lt, "import ") {
		return true
	}
	return false
}
