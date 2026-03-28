package dispatch

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/command"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/edit"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/navigation"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/undo"
)

// executableDispatch turns one validated [intents.ExecutableIntent] payload into an
// [ExecutableResult]. Only [editExecutable] uses [Handler.engine] and [HandleInput.EditCtx];
// other kinds ignore h and in.
type executableDispatch interface {
	dispatch(h *Handler, in HandleInput) (ExecutableResult, error)
}

func executableFor(ex *intents.ExecutableIntent) (executableDispatch, error) {
	switch ex.Kind {
	case intents.ExecutableIntentKindEdit:
		return editExecutable{intent: *ex.Edit}, nil
	case intents.ExecutableIntentKindCommand:
		return commandExecutable{intent: *ex.Command}, nil
	case intents.ExecutableIntentKindNavigate:
		return navigateExecutable{intent: *ex.Navigate}, nil
	case intents.ExecutableIntentKindUndo:
		return undoExecutable{intent: *ex.Undo}, nil
	default:
		return nil, fmt.Errorf("unknown executable intent kind %q", ex.Kind)
	}
}

type editExecutable struct {
	intent intents.EditIntent
}

func (e editExecutable) dispatch(h *Handler, in HandleInput) (ExecutableResult, error) {
	res, err := edit.Dispatch(h.engine, in.EditCtx, e.intent)
	if err != nil {
		return ExecutableResult{}, fmt.Errorf("intent edit: %w", err)
	}
	return ExecutableResult{EditDirective: &res}, nil
}

type commandExecutable struct {
	intent intents.CommandIntent
}

func (c commandExecutable) dispatch(_ *Handler, _ HandleInput) (ExecutableResult, error) {
	res, err := command.Dispatch(c.intent)
	if err != nil {
		return ExecutableResult{}, fmt.Errorf("intent command: %w", err)
	}
	return ExecutableResult{CommandDirective: &res}, nil
}

type navigateExecutable struct {
	intent intents.NavigationIntent
}

func (n navigateExecutable) dispatch(_ *Handler, _ HandleInput) (ExecutableResult, error) {
	res, err := navigation.Dispatch(n.intent)
	if err != nil {
		return ExecutableResult{}, fmt.Errorf("intent navigate: %w", err)
	}
	return ExecutableResult{NavigationDirective: &res}, nil
}

type undoExecutable struct {
	intent intents.UndoIntent
}

func (u undoExecutable) dispatch(_ *Handler, _ HandleInput) (ExecutableResult, error) {
	res, err := undo.Dispatch(u.intent)
	if err != nil {
		return ExecutableResult{}, fmt.Errorf("intent undo: %w", err)
	}
	return ExecutableResult{UndoDirective: &res}, nil
}
