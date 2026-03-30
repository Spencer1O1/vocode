package agent

import (
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
)

// TurnKind discriminates one model completion for the transcript executor.
type TurnKind string

const (
	TurnIrrelevant    TurnKind = "irrelevant"
	TurnFinish        TurnKind = "done"
	TurnGatherContext TurnKind = "request_context"
	TurnIntents       TurnKind = "intents"
)

// TurnResult is exactly one variant: irrelevant, finish, gather context, or a non-empty intent batch.
// Planner-only fields (finish summary, gather spec) are not [intents.Intent] values.
type TurnResult struct {
	Kind             TurnKind
	IrrelevantReason string
	FinishSummary    string
	GatherContext    *agentcontext.GatherContextSpec
	Intents          []intents.Intent
}

// Validate checks the turn union invariant and nested intents.
func (t TurnResult) Validate() error {
	switch t.Kind {
	case TurnIrrelevant:
		if t.GatherContext != nil || len(t.Intents) > 0 || strings.TrimSpace(t.FinishSummary) != "" {
			return fmt.Errorf("agent turn: irrelevant must not set other fields")
		}
	case TurnFinish:
		if t.GatherContext != nil || len(t.Intents) > 0 || strings.TrimSpace(t.IrrelevantReason) != "" {
			return fmt.Errorf("agent turn: finish must not set other fields")
		}
		if err := ValidateFinishSummary(t.FinishSummary); err != nil {
			return fmt.Errorf("agent turn: %w", err)
		}
	case TurnGatherContext:
		if t.GatherContext == nil || len(t.Intents) > 0 ||
			strings.TrimSpace(t.IrrelevantReason) != "" || strings.TrimSpace(t.FinishSummary) != "" {
			return fmt.Errorf("agent turn: gather_context requires requestContext only")
		}
		if err := t.GatherContext.Validate(); err != nil {
			return err
		}
	case TurnIntents:
		if t.GatherContext != nil || strings.TrimSpace(t.IrrelevantReason) != "" || strings.TrimSpace(t.FinishSummary) != "" {
			return fmt.Errorf("agent turn: intents must not set planner-only fields")
		}
		if len(t.Intents) == 0 {
			return fmt.Errorf("agent turn: intents must be non-empty")
		}
		for i := range t.Intents {
			if err := t.Intents[i].Validate(); err != nil {
				return fmt.Errorf("agent turn: intents[%d]: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("agent turn: unknown kind %q", t.Kind)
	}
	return nil
}
