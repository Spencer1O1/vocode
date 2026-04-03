package router

import (
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

// Result is the classifier output for the given flow: route id plus optional structured fields.
type Result struct {
	Flow  flows.ID
	Route string
	// SearchQuery is the primary search string: for workspace_select, prefer the symbol/identifier name
	// (LSP + ripgrep); for select_file, a path/filename fragment. Must be non-empty when those routes are chosen.
	SearchQuery string
	// SearchSymbolKind is an optional classifier hint for workspace_select only (LSP SymbolKind filter on the host).
	// Empty means no kind filter. Ignored for select_file.
	SearchSymbolKind string
}

func (r Result) Validate() error {
	if r.Flow != flows.Root && r.Flow != flows.WorkspaceSelect && r.Flow != flows.SelectFile {
		return fmt.Errorf("flow router: unknown flow %q", r.Flow)
	}
	if err := flows.ValidateRoute(r.Flow, r.Route); err != nil {
		return err
	}
	switch r.Route {
	case "workspace_select", "select_file":
		if strings.TrimSpace(r.SearchQuery) == "" {
			return fmt.Errorf("flow router: route %q requires non-empty search_query", r.Route)
		}
	}
	return nil
}
