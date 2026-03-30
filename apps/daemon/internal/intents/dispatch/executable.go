package dispatch

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/command"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/edit"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/navigation"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/undo"
)

// executableDispatch turns one validated [intents.Intent] into a [Directive].
// Only [editExecutable] uses [Handler.engine] and [HandleInput.EditCtx] (active file text, roots, instruction).
// Other kinds use [HandleInput.Params] only as needed; they ignore [EditCtx].
type executableDispatch interface {
	dispatch(h *Handler, in HandleInput) (Directive, error)
}

func executableFor(ex *intents.Intent) (executableDispatch, error) {
	switch ex.Kind {
	case intents.IntentKindEdit:
		return editExecutable{intent: *ex.Edit}, nil
	case intents.IntentKindCommand:
		return commandExecutable{intent: *ex.Command}, nil
	case intents.IntentKindNavigate:
		return navigateExecutable{intent: *ex.Navigate}, nil
	case intents.IntentKindUndo:
		return undoExecutable{intent: *ex.Undo}, nil
	default:
		return nil, fmt.Errorf("unknown intent kind %q", ex.Kind)
	}
}

type editExecutable struct {
	intent intents.EditIntent
}

func (e editExecutable) dispatch(h *Handler, in HandleInput) (Directive, error) {
	res, err := edit.Dispatch(h.engine, in.EditCtx, e.intent)
	if err != nil {
		return Directive{}, fmt.Errorf("intent edit: %w", err)
	}
	return directiveEdit(&res), nil
}

type commandExecutable struct {
	intent intents.CommandIntent
}

func (c commandExecutable) dispatch(_ *Handler, _ HandleInput) (Directive, error) {
	res, err := command.Dispatch(c.intent)
	if err != nil {
		return Directive{}, fmt.Errorf("intent command: %w", err)
	}
	return directiveCommand(&res), nil
}

type navigateExecutable struct {
	intent intents.NavigationIntent
}

func (n navigateExecutable) dispatch(_ *Handler, _ HandleInput) (Directive, error) {
	res, err := navigation.Dispatch(n.intent)
	if err != nil {
		return Directive{}, fmt.Errorf("intent navigate: %w", err)
	}
	return directiveNavigate(&res), nil
}

type undoExecutable struct {
	intent intents.UndoIntent
}

func (u undoExecutable) dispatch(_ *Handler, _ HandleInput) (Directive, error) {
	res, err := undo.Dispatch(u.intent)
	if err != nil {
		return Directive{}, fmt.Errorf("intent undo: %w", err)
	}
	return directiveUndo(&res), nil
}
