package search

// Path matching in this package combines two ideas:
//
//   - Substring discovery (file_select): case-fold the haystack and needle, then strings.Contains.
//     See pathContainsFold and [PathFragmentMatches].
//
//   - Segment resolution (e.g. move target): walk each path component and match exact name, then
//     EqualFold, then [NormalizePathTokenForMatch] so spoken "my utils" can align with my_utils.
//     See [ResolveWorkspaceRelativePath].
//
// Do not assume one helper fits both: substring search is intentionally loose across the full path;
// segment resolution is stricter (one directory entry per step).

import (
	"strings"
	"unicode"
)

// NormalizePathTokenForMatch folds a path segment for spoken/typed matching: lowercases, then drops
// spaces, underscores, and hyphens. So "my utils", "my_utils", "MyUtils", and "my-utils" align.
func NormalizePathTokenForMatch(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || r == '_' || r == '-' {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		// keep other runes that might appear in folder names (e.g. @scope)
		b.WriteRune(r)
	}
	return b.String()
}

// pathContainsFold reports whether haystack contains needleLower as a substring; needleLower must
// already be lowercased (callers typically lower the user fragment once per search).
func pathContainsFold(haystack, needleLower string) bool {
	if needleLower == "" {
		return false
	}
	return strings.Contains(strings.ToLower(haystack), needleLower)
}
