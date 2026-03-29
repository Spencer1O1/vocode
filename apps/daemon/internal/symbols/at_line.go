package symbols

import (
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/symbols/tags"
)

// ResolveInnermostAtLine implements [Resolver] for [TreeSitterResolver] using [tags] geometry.
func (r *TreeSitterResolver) ResolveInnermostAtLine(workspaceRoot, activeFile string, line0Based, byteCol0 int) (SymbolRef, bool) {
	_ = strings.TrimSpace(workspaceRoot) // reserved for workspace-relative queries later
	if strings.TrimSpace(r.binaryPath) == "" {
		return SymbolRef{}, false
	}
	path := strings.TrimSpace(activeFile)
	if path == "" || line0Based < 0 || byteCol0 < 0 {
		return SymbolRef{}, false
	}
	tagList, err := tags.LoadTags(r.binaryPath, path)
	if err != nil || len(tagList) == 0 {
		return SymbolRef{}, false
	}
	best, ok := tags.SelectInnermostTag(tagList, line0Based, byteCol0)
	if !ok {
		return SymbolRef{}, false
	}
	return tagToSymbolRef(best), true
}
