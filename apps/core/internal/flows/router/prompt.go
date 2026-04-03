package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

// ClassifierSystem builds the system prompt for flow route classification.
// flows.Execution policy is host metadata only; it must never appear here (or in user JSON / schema).
func ClassifierSystem(flow flows.ID) string {
	spec := flows.SpecFor(flow)
	var b strings.Builder
	b.WriteString(strings.TrimSpace(spec.Intro))
	b.WriteString("\n\nRoutes:\n")
	for _, r := range spec.Routes {
		b.WriteString(fmt.Sprintf("- %s: %s\n", r.ID, strings.TrimSpace(r.Description)))
	}
	b.WriteString(`
Return exactly ONE JSON object:
{ "route": "<one of the route ids above>", "search_query": "<string or empty>", "search_symbol_kind": "<string or empty>" }

Rules:
- For "workspace_select", set "search_query" to the primary symbol or identifier name only (e.g. deltaTime, parseHeader, MyClass) — not a prose phrase like "delta time". The host runs LSP workspace symbol search with case-tolerant matching and falls back to ripgrep using derived literals.
  - Exception — literal text search: user gave an exact phrase, error line, log snippet, comment text, or quoted string to find verbatim in files → put that substring in "search_query" (strip outer quotes only) and omit "search_symbol_kind".
  - Optional "search_symbol_kind" (workspace_select only): when you know what kind of symbol they mean, set one of: function, method, class, variable, constant, interface, enum, property, field, constructor, module, struct, type. Omit or use "" when unsure; never guess if ambiguous.
- For "select_file", set "search_query" to a path or filename fragment (e.g. "test.js", "src/api"); set "search_symbol_kind" to "".
- For "workspace_select" and "select_file", "search_query" must be non-empty.
- For all other routes, set "search_query" to "" and "search_symbol_kind" to "".
- No other keys. No markdown.
`)
	return strings.TrimSpace(b.String())
}

// ClassifierUserJSON is the minimal user payload for route classification (flow + utterance).
func ClassifierUserJSON(in Context) ([]byte, error) {
	type payload struct {
		Flow        flows.ID `json:"flow"`
		Instruction string   `json:"instruction"`
	}
	p := payload{
		Flow:        in.Flow,
		Instruction: strings.TrimSpace(in.Instruction),
	}
	return json.MarshalIndent(p, "", "  ")
}

// ClassifierResponseJSONSchema is the JSON Schema for classification (passed to the model client).
func ClassifierResponseJSONSchema(flow flows.ID) map[string]any {
	routes := flows.SpecFor(flow).RouteIDs()
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"route": map[string]any{
				"type": "string",
				"enum": routes,
			},
			"search_query": map[string]any{
				"type": "string",
				"description": "workspace_select: symbol/identifier name or exact literal substring to find in file contents. select_file: path/filename fragment. Otherwise empty.",
			},
			"search_symbol_kind": map[string]any{
				"type": "string",
				"description": "workspace_select only: optional LSP kind hint — function, method, class, variable, constant, interface, enum, property, field, constructor, module, struct, or type. Empty when unknown or for select_file.",
			},
		},
		"required":             []string{"route", "search_query", "search_symbol_kind"},
		"additionalProperties": false,
	}
}
