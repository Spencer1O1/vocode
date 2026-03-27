package symbols

type SymbolRef struct {
	ID   string
	Name string
	Path string
	Line int
	Kind string
}

type Resolver interface {
	ResolveSymbol(workspaceRoot, symbolName, symbolKind, hintPath string) ([]SymbolRef, error)
}
