package flows

// Route is one transcript resolution option within a flow.
type Route struct {
	ID          string
	Description string
}

// Spec is the classifier contract for a flow (prompt + allowed route ids).
type Spec struct {
	Intro  string
	Routes []Route
}

// SpecFor returns the classifier spec for the given flow.
func SpecFor(f ID) Spec {
	switch f {
	case Select:
		return selectSpec()
	case SelectFile:
		return selectFileSpec()
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
	// Routes the select instruction to the select handler (to resolve hit selections), which opens the "select" flow.
	{ID: "select", Description: "Find text or symbols in the workspace."},
	// Routes the select file instruction to the select file handler (to resolve the hit selections), which opens the "select file" flow.
	{ID: "select_file", Description: "Search the workspace for files and folders."},
	// Routes the control instruction to the shared flow control handler. (All flows should implement these controls. Right now only "exit")
	{ID: "control", Description: "Flow navigation (such as exit)"},
	// Routes the irrelevant instruction to the irrelevant handler. (Right now the irrelevant handler ignores the utterance and allows it to show up in the ui under a "skipped" section.)
	{ID: "irrelevant", Description: "Not actionable in this flow."},
}

func rootSpec() Spec {
	rootRoutes := []Route{
		// Routes the question to the question handler (to answer the question)
		{ID: "question", Description: "User asks a question (not a command)."},
	}
	return Spec{
		Intro:  "You are Vocode's classifier for the ROOT flow.\n\nThe user is NOT in a sub-flow. Given one voice transcript, choose exactly one route id. You only classify — details are resolved later.",
		Routes: append(globalRoutes, rootRoutes...),
	}
}

func selectSpec() Spec {
	selectRoutes := []Route{
		// Routes the select control instruction to the select control handler (to navigate the hit list and change the selection)
		{ID: "select_control", Description: "Navigate the hit list (next/previous, jump/goto by number)."},
		// Routes the edit instruction to the edit handler (ai determines the correct edit in the selection which then gets applied)
		{ID: "edit", Description: "They want to edit or change code (scoped edit), not just navigate the list."},
		// Routes the delete instruction to the edit or delete handler (deletes the selection)
		{ID: "delete", Description: "They want to delete this selection."},
	}
	return Spec{
		Intro:  "You are Vocode's classifier for the SELECT result flow.\nThe user already has a list of search hits. Choose exactly one route id. You only classify — details are resolved later.",
		Routes: append(globalRoutes, selectRoutes...),
	}
}

func selectFileSpec() Spec {
	selectFileRoutes := []Route{
		// Routes the select file control instruction to the select file control handler (to navigate the selected file/folder hit list and change the selection)
		{ID: "select_file_control", Description: "Navigate the selected file/folder hit list (next/previous, jump/goto by number)."},
		// Routes the open instruction to the open handler (to open the selected file)
		{ID: "open", Description: "Open the selected file."},
		// Routes the move instruction to the move handler (to move the selected file or folder to a new location)
		{ID: "move", Description: "Move selected file or folder to a new location."},
		// Routes the rename instruction to the rename handler (to rename the selected file or folder)
		{ID: "rename", Description: "Rename selected file or folder."},
		// Routes the create instruction to the create handler (to create a new file or folder in the selected folder)
		{ID: "create", Description: "Create a new file or folder in selected folder."},
		// Routes the delete instruction to the delete handler (to delete the selected file or folder)
		{ID: "delete", Description: "Delete the selected file or folder."},
	}
	return Spec{
		Intro:  "You are Vocode's classifier for the SELECT FILE result flow.\nThe user already has a list of search hits (files and folders). Choose exactly one route id. You only classify — details are resolved later.",
		Routes: append(globalRoutes, selectFileRoutes...),
	}
}
