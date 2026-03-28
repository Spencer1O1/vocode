package dispatch

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/edit"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/requestcontext"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Handler routes planner intents: control intents (done / request_context) vs executables
// (edit / command / navigate / undo → protocol directives).
type Handler struct {
	engine  *edit.Engine
	request *requestcontext.Provider
}

func NewHandler(editEngine *edit.Engine, request *requestcontext.Provider) *Handler {
	return &Handler{engine: editEngine, request: request}
}

// ExecutableResult holds at most one populated directive pointer from a successful executable dispatch.
type ExecutableResult struct {
	EditDirective       *protocol.EditDirective
	CommandDirective    *protocol.CommandDirective
	NavigationDirective *protocol.NavigationDirective
	UndoDirective       *protocol.UndoDirective
}

// DoneResult is the control outcome when the planner stops (done intent). It carries no
// protocol directives; fields may grow (e.g. a host-facing summary string).
type DoneResult struct{}

// RequestContextFulfilled is the control outcome after a request_context intent is fulfilled.
type RequestContextFulfilled struct {
	PlanningContext agent.PlanningContext
}

// ControlResult is exactly one of [DoneResult] or [RequestContextFulfilled] (union).
type ControlResult struct {
	Done      *DoneResult
	Fulfilled *RequestContextFulfilled
}

// HandleOutcome is the result of [Handler.Handle]: either a [ControlResult] or an [ExecutableResult].
type HandleOutcome struct {
	Control    *ControlResult
	Executable *ExecutableResult
}

// HandleInput is everything one [Handler.Handle] call needs: transcript params, planner
// context snapshot, the validated intent union, and (for executable intents) edit execution
// state built by the transcript executor. Control vs executable branches read different fields;
// unused fields are intentionally ignored (e.g. done ignores all of this; edit uses Engine + EditCtx).
type HandleInput struct {
	Params  protocol.VoiceTranscriptParams
	TurnCtx agent.PlanningContext
	Intent  intents.Intent
	EditCtx edit.EditExecutionContext
}

// Handle validates the union and dispatches control intents vs executables.
func (h *Handler) Handle(in HandleInput) (HandleOutcome, error) {
	if err := in.Intent.Validate(); err != nil {
		return HandleOutcome{}, err
	}
	if c := in.Intent.Control; c != nil {
		return h.dispatchControl(c, in)
	}
	ex := in.Intent.Executable
	if ex == nil {
		return HandleOutcome{}, fmt.Errorf("planner intent: missing executable")
	}
	return h.dispatchExecutable(ex, in)
}

func (h *Handler) dispatchControl(c *intents.ControlIntent, in HandleInput) (HandleOutcome, error) {
	op, err := controlFor(c)
	if err != nil {
		return HandleOutcome{}, err
	}
	cr, err := op.dispatch(h, in)
	if err != nil {
		return HandleOutcome{}, err
	}
	return HandleOutcome{Control: cr}, nil
}

func (h *Handler) dispatchExecutable(ex *intents.ExecutableIntent, in HandleInput) (HandleOutcome, error) {
	op, err := executableFor(ex)
	if err != nil {
		return HandleOutcome{}, err
	}
	dr, err := op.dispatch(h, in)
	if err != nil {
		return HandleOutcome{}, err
	}
	return HandleOutcome{Executable: &dr}, nil
}
