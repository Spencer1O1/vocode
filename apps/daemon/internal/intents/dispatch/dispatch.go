// Package dispatch routes validated executable [intents.Intent] values to protocol-shaped results.
//
// [Handler] holds the edit engine; [HandleInput] carries transcript params, the validated [intents.Intent],
// and edit mechanical context ([edit.EditExecutionContext]). Daemon-enriched IDE state from gather rounds
// is not part of dispatch—if a directive strategy needs it later, add an explicit field (do not overload “gather”).
// Turn-level gather-context and finish are orchestrated by the transcript executor (see [gather] package), not here.
package dispatch

import (
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/edit"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Handler is the dispatch context: shared dependencies for executable intent strategies.
type Handler struct {
	engine *edit.Engine
}

func NewHandler(editEngine *edit.Engine) *Handler {
	return &Handler{engine: editEngine}
}

// HandleInput is per-call dispatch context for one executable intent.
type HandleInput struct {
	Params  protocol.VoiceTranscriptParams
	Intent  intents.Intent
	EditCtx edit.EditExecutionContext
}

// Handle validates and dispatches one executable intent.
func (h *Handler) Handle(in HandleInput) (Directive, error) {
	if err := in.Intent.Validate(); err != nil {
		return Directive{}, err
	}
	op, err := executableFor(&in.Intent)
	if err != nil {
		return Directive{}, err
	}
	return op.dispatch(h, in)
}
