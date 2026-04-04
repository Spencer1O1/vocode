package workspaceselectflow

import "strings"

func isJsTsSkippableLeadingDirectiveLine(line string) bool {
	t := strings.TrimSpace(line)
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
	return false
}

func isJsTsImportOrReexportFromLine(line string) bool {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "import ") && !strings.HasPrefix(t, "import(") {
		return true
	}
	if strings.HasPrefix(t, "export ") && strings.Contains(t, " from ") {
		return true
	}
	return false
}

// jsTsImportInsertLine returns the 0-based line index for a zero-width insert of new import lines.
// Imports are placed immediately after the contiguous top import / re-export-from block. Blank lines
// after that block no longer push the insert index down to the next statement (which used to split
// new imports from existing ones with an extra gap and leave them flush against following code).
func jsTsImportInsertLine(lines []string) int {
	i := 0
	for i < len(lines) {
		if !isJsTsSkippableLeadingDirectiveLine(lines[i]) {
			break
		}
		i++
	}

	lastImport := -1
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			break
		}
		if isJsTsImportOrReexportFromLine(lines[i]) {
			lastImport = i
			i++
			continue
		}
		break
	}
	if lastImport >= 0 {
		return lastImport + 1
	}

	for i < len(lines) {
		if isJsTsSkippableLeadingDirectiveLine(lines[i]) {
			i++
			continue
		}
		return i
	}
	return len(lines)
}

// jstsFinalizeImportBlock appends a newline so one blank line separates the new import block from
// following top-level code when that line is not another import/re-export-from.
// If insertLine already sits on a blank line before that code, the file already has the gap — do not
// append another newline or organize/edit flows end up with two blank lines before export/default.
func jstsFinalizeImportBlock(src []string, insertLine int, block string) string {
	if block == "" {
		return block
	}
	j := insertLine
	hadBlankAfterInsertSlot := false
	for j < len(src) && strings.TrimSpace(src[j]) == "" {
		hadBlankAfterInsertSlot = true
		j++
	}
	if j >= len(src) {
		return block
	}
	if isJsTsImportOrReexportFromLine(src[j]) {
		return block
	}
	if hadBlankAfterInsertSlot {
		return block
	}
	if strings.HasSuffix(block, "\n\n") {
		return block
	}
	return block + "\n"
}
