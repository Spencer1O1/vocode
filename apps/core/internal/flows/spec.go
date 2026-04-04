package flows

// Global route "create" adds new text in the active editor file (placement resolved in a later step).
// Flow file_select also defines "create_entry": a new file or folder on disk from the path list — not editor content.

// Route is one transcript resolution option within a flow.
type Route struct {
	ID          string
	Description string
	// Execution is host-only ordering metadata; never exposed to the routing model.
	Execution Execution
}

// Spec is the classifier contract for a flow (prompt + allowed route ids).
type Spec struct {
	Intro  string
	Routes []Route
}

// SpecFor returns the classifier spec for the given flow.
func SpecFor(f ID) Spec {
	switch f {
	case WorkspaceSelect:
		return workspaceSelectSpec()
	case SelectFile:
		return fileSelectSpec()
	default:
		return rootSpec()
	}
}

// RouteIDs returns route ids in spec order (matches prompt / JSON enum).
func (s Spec) RouteIDs() []string {
	out := make([]string, len(s.Routes))
	for i, r := range s.Routes {
		out[i] = r.ID
	}
	return out
}

var globalRoutes = []Route{
	{ID: "workspace_select", Description: "Find or go to symbols, identifiers, or text inside files (by name or literal substring). Do not use for the verb “open” — you cannot open a function; “open” is file_select. Default when ambiguous: no path-on-disk signal → prefer this over file_select. “Go to main” without file cues → workspace_select (symbol search).", Execution: ExecutionSerialized},
	{ID: "file_select", Description: "Find or open files or folders by basename. Reserve “open” for this route (open the explore file, open package.json). Signals: extension, file/folder/directory/document, basename-shaped token; STT “dot” → period. search_query is one segment — no slashes. Host handles missing workspace if needed.", Execution: ExecutionSerialized},
	{ID: "create", Description: "Add new content to the active editor file. The user must name or clearly imply what to add (function, variable, type, comment, import, test, etc.). Vague “add something” / “put code here” with no identifiable what → not create. Placement is optional in speech; a later step infers where to insert.", Execution: ExecutionSerialized},
	{ID: "command", Description: "Run terminal/shell work now: install, scaffold, git init, run tests/build, dev server, etc. Use for clear execute-now intent, including polite questions like “can you run the tests?” when they mean execution, not an explanation.", Execution: ExecutionSerialized},
	{ID: "control", Description: "Dismiss or leave the current flow only: exit, cancel, go back, stop, quit, never mind, and short synonyms.", Execution: ExecutionImmediate},
	{ID: "irrelevant", Description: "Not actionable in this flow, off-topic, talking to someone else, noise, or nonsensical. Also: thanks/okay/got it (not clearly exit); vague create with no “what”; ROOT + selection + only “fix this”/“make it work” with no named what to add.", Execution: ExecutionImmediate},
}

func rootSpec() Spec {
	rootRoutes := []Route{
		{ID: "question", Description: "Informational intent only: how/what/why, explanations. Not when the user clearly wants execution now — that is command even if phrased as a question.", Execution: ExecutionImmediate},
	}
	return Spec{
		Intro: `You are Vocode's classifier for the ROOT flow. Input is speech-to-text; expect informal phrasing.

The user is not in a sub-flow. User JSON may include activeFile, hasNonemptySelection, workspaceRoot, hostPlatform, workspaceFolderOpen.

Tie-breaks (ROOT):
- question: informational / who-what-when-where-why-how only. command: imperative or clear execute-now intent (including “can you run …?” when they mean run it, not explain it).
- workspace_select vs file_select: Verbs find, go to, select, look for can mean either; use file signals (extension, file/folder/directory/document, obvious basename) → file_select. The verb “open” always favors file_select (you open files/folders, not symbols). If ambiguous and no path signal and no “open”, default workspace_select. “Go to main” without file/open cues → workspace_select.
- Compound utterance (search + create/command in one line): prefer workspace_select or file_select over create or command (search wins).
- create: only when the user names or clearly implies what to add. Vague “add something” with no what → irrelevant.
- control: dismiss/leave the flow. thanks / okay / got it (not clearly exit) → irrelevant, not control.
- irrelevant: ROOT + non-empty selection + only vague “fix this” / “make it work” with no named what → irrelevant (not create).

For file_select, never put a full path in search_query — basename only (e.g. game.js).

Choose exactly one route. You only classify; details are resolved later.`,
		Routes: append(globalRoutes, rootRoutes...),
	}
}

func workspaceSelectSpec() Spec {
	wsRoutes := []Route{
		{ID: "workspace_select_control", Description: "Navigate the existing workspace hit list only: next, previous, pick by position (first/second hit, third result, short \"go to two\"). Not for go-to plus a symbol or file name—that is workspace_select with search_query.", Execution: ExecutionImmediate},
		{ID: "edit", Description: "Change code at the current focus or selection. When hasNonemptySelection is true and they say vague “fix this” / “make it work” (improving existing code, not naming new content to add), prefer edit — not irrelevant.", Execution: ExecutionSerialized},
		{ID: "rename", Description: "Rename the thing at the current hit or selection (e.g. rename X to Y, call it Z).", Execution: ExecutionSerialized},
		{ID: "delete", Description: "Delete the current selection or hit.", Execution: ExecutionSerialized},
	}
	return Spec{
		Intro: `You are Vocode's classifier for the WORKSPACE SELECT flow. The user has workspace search hits; the editor may have a non-empty selection. Input is speech-to-text.

User JSON may include hasNonemptySelection and activeFile.

When starting a new search in this flow, use workspace_select with a non-empty search_query. "Go to" plus a symbol or component name is workspace_select; "open" plus a name is file_select (not workspace_select). Neither is list control. For list navigation only (next, previous, pick Nth hit), use workspace_select_control.

Choose exactly one route. You only classify; details are resolved later.`,
		Routes: append(globalRoutes, wsRoutes...),
	}
}

func fileSelectSpec() Spec {
	fsRoutes := []Route{
		{ID: "file_select_control", Description: "Navigate the file hit list (next, previous, pick by number, etc.).", Execution: ExecutionImmediate},
		{ID: "move", Description: "Move the selected file or folder to another path.", Execution: ExecutionSerialized},
		{ID: "rename", Description: "Rename the selected file or folder.", Execution: ExecutionSerialized},
		{ID: "create_entry", Description: "New file or folder on disk under the selected row. search_query must be empty.", Execution: ExecutionSerialized},
		{ID: "delete", Description: "Delete the selected file. (Workspace root and folders are not deletable via this route.)", Execution: ExecutionSerialized},
	}
	return Spec{
		Intro: `You are Vocode's classifier for the SELECT FILE flow. The user has file/folder path hits. Input is speech-to-text.

workspace_select: search inside file contents (not “open”). file_select: basename path lookup and “open …” for files/folders.

create_entry: new path on disk under the selection — search_query must be "". create: editor buffer only, not new disk path from this flow.

Choose exactly one route. You only classify; details are resolved later.`,
		Routes: append(globalRoutes, fsRoutes...),
	}
}
