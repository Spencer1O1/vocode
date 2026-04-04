package router

import (
	"strings"
	"unicode"
)

// DisambiguateClassifierResult forces a stable split between file path search and workspace search.
// Path discourse (file / folder / directory, or "open" for path browse) → file_select; otherwise
// workspace_select. The model still supplies search_query and optional search_symbol_kind for workspace.
func DisambiguateClassifierResult(in Context, res Result) Result {
	if res.Route != "file_select" && res.Route != "workspace_select" {
		return res
	}
	u := strings.TrimSpace(in.Instruction)
	if u == "" {
		return res
	}
	if utteranceImpliesPathDiscourse(u) {
		res.Route = "file_select"
		res.SearchSymbolKind = ""
		return res
	}
	res.Route = "workspace_select"
	res.SearchQuery = NormalizeWorkspaceSelectSearchQuery(res.SearchQuery)
	return res
}

func utteranceHasWholeWord(q, w string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	w = strings.ToLower(strings.TrimSpace(w))
	if q == "" || w == "" {
		return false
	}
	for _, tok := range strings.FieldsFunc(q, func(r rune) bool {
		return r == '_' || unicode.IsPunct(r) || unicode.IsSpace(r)
	}) {
		if tok == w {
			return true
		}
	}
	return false
}

func utteranceImpliesPathDiscourse(u string) bool {
	for _, w := range []string{
		"file", "folder", "folders", "directory", "directories",
		"open", // "open app", "open readme" — path browse without saying "file"
	} {
		if utteranceHasWholeWord(u, w) {
			return true
		}
	}
	return false
}
