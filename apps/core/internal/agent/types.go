package agent

// EditorSnapshot is editor context passed into flow classification.
type EditorSnapshot struct {
	ActiveFilePath string
	WorkspaceRoot  string
	CursorSymbol   *SymbolRef
}

// SymbolRef is a lightweight cursor symbol reference.
type SymbolRef struct {
	Name string
	Kind string
}
