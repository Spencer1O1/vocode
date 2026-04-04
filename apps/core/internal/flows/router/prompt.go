package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

// ClassifierSystem builds the system prompt for flow route classification.
// It starts from flows.Spec (intro + per-route descriptions), then appends shared JSON output Rules.
// Flow-specific tie-break bullets are appended below (workspace select, select file); workspace select uses classifier user JSON.
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
- For "workspace_select", set "search_query" to the primary symbol or identifier name only (e.g. deltaTime, parseHeader, MyClass) — not a prose phrase like "delta time".
  - Exception — literal text search: user gave an exact phrase, error line, log snippet, comment text, or quoted string to find verbatim in files → put that substring in "search_query" (strip outer quotes only) and omit "search_symbol_kind".
  - Optional "search_symbol_kind" (workspace_select only): when you know what kind of symbol they mean, set one of: function, method, class, variable, constant, interface, enum, property, field, constructor, module, struct, type. Omit or use "" when unsure; never guess if ambiguous.
- For "file_select", set "search_query" to a single file or folder name only (basename): e.g. game.js, README.md, Res. No slashes, no path segments, no absolute paths — never paste activeFile or any full system path. STT may say "dot" for a period in the name — rewrite to real punctuation in that one segment only. Set "search_symbol_kind" to "". Still use file_select if workspaceFolderOpen is false; the host handles missing workspace later.
- For "workspace_select" and "file_select", "search_query" must be non-empty.
- For all other routes, set "search_query" to "" and "search_symbol_kind" to "".
- "question" vs "command": "question" = informational / how-what-why. "command" = imperative or clear execute-now intent for terminal work (install, run tests, scaffold, git), including "can you run …?" / "could you npm install?" when they mean execution, not a lesson.
- "workspace_select" vs "file_select": Verbs find, go to, navigate to, select, look for can mean either — use path signals (extension, file/folder/directory/document, obvious basename) → "file_select". The verb "open" is reserved for files and folders: "open …" → "file_select" with basename in search_query, never "workspace_select" for a symbol hunt. If ambiguous and no path signal and no "open", default "workspace_select". "Go to main" without file/open cues → "workspace_select".
- "go to" + named symbol or component, or file-target phrase: use "workspace_select" or "file_select" respectively; search_query = identifier or basename only (strip find, go, to, the). "open" + name → always "file_select". Do NOT use "workspace_select_control" for those — control is only for the existing hit list (next, previous, first/second hit/result, bare short picks).
- Compound utterance (e.g. find X and then add Y): if both search and create/command apply, prefer "workspace_select" or "file_select" over "create" or "command" (search wins for this turn).
- "create" gate: use "create" only when the user names or clearly implies what to add (function, comment, import, type, etc.). Vague "add something" / "put code here" with no identifiable what → "irrelevant", not "create".
- "control" vs "irrelevant": "control" = dismiss/leave the flow (exit, cancel, stop, quit, go back, never mind). Casual "thanks" / "okay" / "got it" without clear exit → "irrelevant".
- ROOT flow + hasNonemptySelection + only vague "fix this" / "make it work" with no named what to add → "irrelevant", not "create".
- No other keys. No markdown.
`)
	if flow == flows.WorkspaceSelect {
		b.WriteString(`

Workspace select — user JSON may include hasNonemptySelection (true when the editor selection is non-empty) and activeFile:
- When hasNonemptySelection is true and the utterance is vague "fix this" / "make it work" / improve existing code without naming new content to add, prefer "edit" over "irrelevant" or "workspace_select".
- When hasNonemptySelection is true, if the utterance matches global "create" or "rename" per the route list, prefer those when appropriate.
- When hasNonemptySelection is true, the utterance is an imperative to change existing code and they are not starting a new workspace search, prefer "edit" over "workspace_select".
- Starting a new search while in this flow → "workspace_select" with a fresh non-empty search_query. List navigation only → "workspace_select_control".
- Phrases like "go to the TabTwo screen" → workspace_select; "open the explore file" / "go to the explore file" with file cue → file_select — new search, not workspace_select_control. Put symbol or basename in search_query, not list position.
`)
	}
	if flow == flows.SelectFile {
		b.WriteString(`

Select file — global "create" vs "create_entry": user JSON flow is file_select; activeFile may still be set from the editor. Use create_entry when they name a new file or folder on disk under the list row (add, make, create, new + a name; STT may say "dot" for "."). For create_entry set search_query to "". Use create only for changing the open editor buffer, not for a new disk path from this flow.
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
	switch flow {
	case flows.Root:
		return "Non-empty only for workspace_select and file_select. For question, command, create, control, irrelevant: empty. Vague create (no what) → route irrelevant with empty search fields."
	case flows.WorkspaceSelect:
		return "Non-empty for workspace_select and file_select only. For workspace_select_control, edit, rename, delete, command, create, control, irrelevant: empty."
	case flows.SelectFile:
		return "Non-empty for workspace_select and file_select only. For file_select_control, move, rename, create_entry (must be empty), delete, command, create, control, irrelevant: empty."
	default:
		return "Non-empty only for workspace_select and file_select; otherwise empty."
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
