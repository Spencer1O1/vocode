package actionplan

import (
	"fmt"
	"strings"
)

// NextActionKind is the iterative planner response discriminant.
type NextActionKind string

const (
	NextActionKindEdit           NextActionKind = "edit"
	NextActionKindRunCommand     NextActionKind = "run_command"
	NextActionKindNavigate       NextActionKind = "navigate"
	NextActionKindRequestContext NextActionKind = "request_context"
	NextActionKindDone           NextActionKind = "done"
)

// NextAction is one model turn output in iterative planning mode.
type NextAction struct {
	Kind NextActionKind `json:"kind"`

	Edit           *EditIntent       `json:"edit,omitempty"`
	RunCommand     *CommandIntent    `json:"runCommand,omitempty"`
	Navigate       *NavigationIntent `json:"navigate,omitempty"`
	RequestContext *RequestContextIntent `json:"requestContext,omitempty"`
}

func ValidateNextAction(a NextAction) error {
	switch a.Kind {
	case NextActionKindEdit:
		if a.Edit == nil {
			return fmt.Errorf("next action: kind %q requires edit", a.Kind)
		}
		return ValidateEditIntent(*a.Edit)
	case NextActionKindRunCommand:
		if a.RunCommand == nil {
			return fmt.Errorf("next action: kind %q requires runCommand", a.Kind)
		}
		if strings.TrimSpace(a.RunCommand.Command) == "" {
			return fmt.Errorf("next action: runCommand.command is empty")
		}
		return nil
	case NextActionKindNavigate:
		if a.Navigate == nil {
			return fmt.Errorf("next action: kind %q requires navigate", a.Kind)
		}
		return ValidateNavigationIntent(*a.Navigate)
	case NextActionKindRequestContext:
		if a.RequestContext == nil {
			return fmt.Errorf("next action: kind %q requires requestContext", a.Kind)
		}
		switch a.RequestContext.Kind {
		case RequestContextKindSymbols, RequestContextKindFileExcerpt, RequestContextKindUsages:
			return nil
		default:
			return fmt.Errorf("next action: unknown requestContext kind %q", a.RequestContext.Kind)
		}
	case NextActionKindDone:
		return nil
	default:
		return fmt.Errorf("next action: unknown kind %q", a.Kind)
	}
}
