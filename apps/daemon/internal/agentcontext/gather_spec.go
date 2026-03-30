package agentcontext

import "fmt"

// GatherContextKind selects how the planner asks to enrich gathered IDE context (no host directive).
type GatherContextKind string

const (
	GatherContextKindSymbols     GatherContextKind = "request_symbols"
	GatherContextKindFileExcerpt GatherContextKind = "request_file_excerpt"
	GatherContextKindUsages      GatherContextKind = "request_usages"
)

// GatherContextSpec is the payload for a turn-level gather-context response (wire: nested under requestContext).
type GatherContextSpec struct {
	Kind      GatherContextKind `json:"kind"`
	Path      string            `json:"path,omitempty"`
	Query     string            `json:"query,omitempty"`
	SymbolID  string            `json:"symbolId,omitempty"`
	MaxResult int               `json:"maxResult,omitempty"`
}

// Validate checks payload constraints for a gather-context turn.
func (g *GatherContextSpec) Validate() error {
	if g == nil {
		return fmt.Errorf("gather_context: nil spec")
	}
	switch g.Kind {
	case GatherContextKindSymbols, GatherContextKindFileExcerpt, GatherContextKindUsages:
		return nil
	default:
		return fmt.Errorf("gather_context: unknown kind %q", g.Kind)
	}
}
