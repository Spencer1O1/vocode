package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

// ClassifierSystem builds the system prompt for flow route classification (spec + JSON rules).
// flows.Execution is host-only and must not appear here.
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
- Workspace vs file (search routes): if they clearly mean a path on disk (words like file, folder, directory, or open + a name), choose "file_select"; otherwise choose "workspace_select" for symbol/text-in-files search. The host may adjust this from the raw utterance; still pick the route that matches intent.
- "workspace_select": search_query = one identifier token or literal substring — strip trailing discourse that only names the kind of thing (component, function, class, method, symbol, hook, …); put kind in search_symbol_kind when you set it, not on the end of search_query. Example: "find the test component" → search_query "test" (not "test component"). Verbatim phrase / log line / quoted text → full substring; omit search_symbol_kind.
- "file_select": search_query = one basename from what they said (strip filler); STT "dot" → period in that name. Never copy activeFile's basename unless they said that filename. No slashes or full paths. search_symbol_kind "".
- "workspace_select" and "file_select" need non-empty search_query; other routes use "" for both search fields.
- "question" vs "command": run-now execution → command even if phrased as a question.
- workspace_select_control / file_select_control: only for next/previous/pick-N on the current list, not a new name to find.
- Compound utterance (search + create/command): search route wins this turn.
- "create" / "control" / "irrelevant": per route descriptions; hasNonemptySelection → not "create", use "edit" for the selection.
- No other keys. No markdown.
`)
	if flow == flows.WorkspaceSelect {
		b.WriteString(`

Workspace select — when hasNonemptySelection is true: vague "fix this" / "make it work" without naming new content → prefer "edit" over "irrelevant" or "workspace_select"; imperative to change existing code without starting a new search → prefer "edit" over "workspace_select".
`)
	}
	return strings.TrimSpace(b.String())
}

// ClassifierUserJSON is the minimal user payload for route classification (flow + utterance).
func ClassifierUserJSON(in Context) ([]byte, error) {
	type payload struct {
		Flow                 flows.ID `json:"flow"`
		Instruction          string   `json:"instruction"`
		ActiveFile           string   `json:"activeFile,omitempty"`
		HasNonemptySelection bool     `json:"hasNonemptySelection,omitempty"`
		WorkspaceRoot        string   `json:"workspaceRoot,omitempty"`
		HostPlatform         string   `json:"hostPlatform,omitempty"`
		WorkspaceFolderOpen  bool     `json:"workspaceFolderOpen,omitempty"`
	}
	p := payload{
		Flow:                 in.Flow,
		Instruction:          strings.TrimSpace(in.Instruction),
		ActiveFile:           strings.TrimSpace(in.ActiveFile),
		HasNonemptySelection: in.HasNonemptySelection,
		WorkspaceRoot:        strings.TrimSpace(in.WorkspaceRoot),
		HostPlatform:         strings.TrimSpace(in.HostPlatform),
		WorkspaceFolderOpen:  in.WorkspaceFolderOpen,
	}
	return json.MarshalIndent(p, "", "  ")
}

func classifierSearchQueryDescription(flow flows.ID) string {
	hint := ` Non-empty for workspace_select/file_select only: one name token (file_select) or identifier/literal (workspace_select); do not echo activeFile unless they said it.`
	switch flow {
	case flows.Root:
		return "Empty except workspace_select and file_select (question/command/create/control/irrelevant use empty search fields; vague create → irrelevant)." + hint
	case flows.WorkspaceSelect:
		return "Empty except workspace_select and file_select." + hint
	case flows.SelectFile:
		return "Empty except workspace_select and file_select; create_entry must use empty search fields." + hint
	default:
		return "Empty except workspace_select and file_select." + hint
	}
}

// ClassifierResponseJSONSchema is the JSON Schema for classification (passed to the model client).
func ClassifierResponseJSONSchema(flow flows.ID) map[string]any {
	routes := flows.SpecFor(flow).RouteIDs()
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"route": map[string]any{
				"type":        "string",
				"enum":        routes,
				"description": "Exactly one route id from the system prompt for this flow. See tie-break rules for question/command, workspace vs file, create gate, compound utterances.",
			},
			"search_query": map[string]any{
				"type":        "string",
				"description": classifierSearchQueryDescription(flow),
			},
			"search_symbol_kind": map[string]any{
				"type":        "string",
				"description": "workspace_select only: optional symbol kind — function, method, class, variable, constant, interface, enum, property, field, constructor, module, struct, type. Empty for file_select or any other route.",
			},
		},
		"required":             []string{"route", "search_query", "search_symbol_kind"},
		"additionalProperties": false,
	}
}
