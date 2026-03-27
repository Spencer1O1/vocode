package agent

import (
	"context"

	"vocoding.net/vocode/v2/apps/daemon/internal/actionplan"
	"vocoding.net/vocode/v2/apps/daemon/internal/symbols"
)

// ModelInput is everything the model needs to propose the next action.
// Fields may grow (active file, selection, workspace roots, etc.).
type ModelInput struct {
	Transcript       string
	CompletedActions []actionplan.NextAction
	Context          PlanningContext
}

type FileExcerpt struct {
	Path    string
	Content string
}

// PlanningContext is bounded context gathered via request_context turns.
type PlanningContext struct {
	Symbols  []symbols.SymbolRef
	Excerpts []FileExcerpt
	Notes    []string
}

// ModelClient is the iterative planning contract.
type ModelClient interface {
	NextAction(ctx context.Context, in ModelInput) (actionplan.NextAction, error)
}
