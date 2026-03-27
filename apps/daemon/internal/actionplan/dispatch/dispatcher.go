package dispatch

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/daemon/internal/actionplan"
	"vocoding.net/vocode/v2/apps/daemon/internal/commandexec"
	"vocoding.net/vocode/v2/apps/daemon/internal/edits"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Dispatcher runs validated steps step-by-step.
type Dispatcher struct {
	edits    *edits.Service
	commands *commandexec.Service
}

func NewDispatcher(editsService *edits.Service, commandService *commandexec.Service) *Dispatcher {
	return &Dispatcher{edits: editsService, commands: commandService}
}

// StepResult is the outcome of executing one NextAction.
type StepResult struct {
	EditResult    *protocol.EditApplyResult
	CommandParams *protocol.CommandRunParams
	Navigation    *actionplan.NavigationIntent
}

func (d *Dispatcher) ExecuteNextAction(next actionplan.NextAction, editCtx edits.EditExecutionContext) (StepResult, error) {
	if err := actionplan.ValidateNextAction(next); err != nil {
		return StepResult{}, err
	}
	switch next.Kind {
	case actionplan.NextActionKindEdit:
		res, err := d.edits.ApplyIntent(editCtx, *next.Edit)
		if err != nil {
			return StepResult{}, fmt.Errorf("next action edit: %w", err)
		}
		return StepResult{EditResult: &res}, nil
	case actionplan.NextActionKindRunCommand:
		params := next.RunCommand.CommandParams()
		if d.commands != nil {
			if err := d.commands.Validate(params); err != nil {
				return StepResult{}, fmt.Errorf("next action run_command: %w", err)
			}
		}
		return StepResult{CommandParams: &params}, nil
	case actionplan.NextActionKindNavigate:
		return StepResult{Navigation: next.Navigate}, nil
	case actionplan.NextActionKindDone:
		return StepResult{}, fmt.Errorf("done is not executable")
	case actionplan.NextActionKindRequestContext:
		return StepResult{}, fmt.Errorf("request_context is not executable")
	default:
		return StepResult{}, fmt.Errorf("unknown next action kind %q", next.Kind)
	}
}
