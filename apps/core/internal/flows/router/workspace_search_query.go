package router

import "strings"

// Trailing words users add when describing *what* to find (component, function, …). They are not
// part of the symbol name; search_symbol_kind covers kind when the model sets it.
// Only stripped when search_query has 2+ whitespace-separated tokens and the last token matches.
var workspaceSearchDiscourseSuffix = map[string]struct{}{
	"component":   {},
	"components":  {},
	"symbol":      {},
	"symbols":     {},
	"function":    {},
	"functions":   {},
	"method":      {},
	"methods":     {},
	"class":       {},
	"classes":     {},
	"hook":        {},
	"hooks":       {},
	"definition":  {},
	"definitions": {},
}

// NormalizeWorkspaceSelectSearchQuery drops trailing discourse tokens (e.g. "test component" → "test").
// Single-token queries are unchanged ("symbol" stays "symbol"). Empty-after-strip falls back to the input.
func NormalizeWorkspaceSelectSearchQuery(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return q
	}
	fields := strings.Fields(q)
	for len(fields) > 1 {
		last := strings.Trim(strings.ToLower(fields[len(fields)-1]), ".,;:!?")
		if _, ok := workspaceSearchDiscourseSuffix[last]; !ok {
			break
		}
		fields = fields[:len(fields)-1]
	}
	out := strings.Join(fields, " ")
	out = strings.TrimSpace(out)
	if out == "" {
		return q
	}
	return out
}
