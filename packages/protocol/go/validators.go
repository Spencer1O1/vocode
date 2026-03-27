package protocol

import (
	"errors"
	"fmt"
	"strings"
)

// EditApplyResult validation lives here alongside future protocol-level validators
// (mirrors typescript/validators.ts conceptually).

func (r EditApplyResult) Validate() error {
	switch r.Kind {
	case "success":
		if r.Actions == nil {
			return errors.New("success result must include actions")
		}
		if r.Failure != nil || r.Reason != "" {
			return errors.New("success result must not contain failure or reason")
		}
	case "failure":
		if r.Failure == nil {
			return errors.New("failure result must include failure")
		}
		if len(r.Actions) > 0 || r.Reason != "" {
			return errors.New("failure result must not contain actions or reason")
		}
	case "noop":
		if r.Reason == "" {
			return errors.New("noop result must include reason")
		}
		if len(r.Actions) > 0 || r.Failure != nil {
			return errors.New("noop result must not contain actions or failure")
		}
	default:
		return errors.New("unknown edit.apply result kind")
	}

	return nil
}

func (s VoiceTranscriptStepResult) Validate() error {
	switch s.Kind {
	case "edit":
		if s.EditResult == nil || s.CommandParams != nil || s.NavigationIntent != nil {
			return errors.New("voice transcript step: kind edit requires editResult and no commandParams/navigationIntent")
		}
		return s.EditResult.Validate()
	case "command":
		if s.CommandParams == nil || s.EditResult != nil || s.NavigationIntent != nil {
			return errors.New("voice transcript step: kind command requires commandParams and no editResult/navigationIntent")
		}
		if strings.TrimSpace(s.CommandParams.Command) == "" {
			return errors.New("voice transcript step: command requires non-empty commandParams.command")
		}
		// CommandRunParams has no additional protocol-level validation yet; host-side
		// policy executes the safety checks.
		return nil
	case "navigate":
		if s.NavigationIntent == nil || s.EditResult != nil || s.CommandParams != nil {
			return errors.New("voice transcript step: kind navigate requires navigationIntent and no editResult/commandParams")
		}
		return validateNavigationIntent(s.NavigationIntent)
	default:
		return fmt.Errorf("voice transcript step: unknown kind %q", s.Kind)
	}
}

func validateNavigationIntent(n *NavigationIntent) error {
	if n == nil {
		return errors.New("voice transcript step: navigate requires navigationIntent")
	}
	kind := strings.TrimSpace(n.Kind)
	if kind == "" {
		return errors.New("voice transcript step: navigate requires non-empty navigationIntent.kind")
	}

	payloads := 0
	if n.OpenFile != nil {
		payloads++
	}
	if n.RevealSymbol != nil {
		payloads++
	}
	if n.MoveCursor != nil {
		payloads++
	}
	if n.SelectRange != nil {
		payloads++
	}
	if n.RevealEdit != nil {
		payloads++
	}
	if payloads != 1 {
		return errors.New("voice transcript step: navigate requires exactly one navigation payload")
	}

	switch kind {
	case "open_file":
		if n.OpenFile == nil || strings.TrimSpace(n.OpenFile.Path) == "" {
			return errors.New("voice transcript step: open_file requires openFile.path")
		}
	case "reveal_symbol":
		if n.RevealSymbol == nil || strings.TrimSpace(n.RevealSymbol.SymbolName) == "" {
			return errors.New("voice transcript step: reveal_symbol requires revealSymbol.symbolName")
		}
	case "move_cursor":
		if n.MoveCursor == nil || n.MoveCursor.Target.Line < 0 || n.MoveCursor.Target.Char < 0 {
			return errors.New("voice transcript step: move_cursor requires moveCursor.target with non-negative line/char")
		}
	case "select_range":
		if n.SelectRange == nil {
			return errors.New("voice transcript step: select_range requires selectRange.target")
		}
		t := n.SelectRange.Target
		if t.StartLine < 0 || t.StartChar < 0 || t.EndLine < 0 || t.EndChar < 0 {
			return errors.New("voice transcript step: select_range requires non-negative target positions")
		}
	case "reveal_edit":
		if n.RevealEdit == nil || strings.TrimSpace(n.RevealEdit.EditId) == "" {
			return errors.New("voice transcript step: reveal_edit requires revealEdit.editId")
		}
	default:
		return fmt.Errorf("voice transcript step: unknown navigation kind %q", kind)
	}

	return nil
}

func (r VoiceTranscriptResult) Validate() error {
	if !r.Accepted {
		return errors.New("voice transcript result must have accepted=true")
	}
	if r.PlanError != "" && len(r.Results) > 0 {
		return errors.New("voice transcript result must not include both planError and results")
	}
	for i := range r.Results {
		if err := r.Results[i].Validate(); err != nil {
			return fmt.Errorf("voice transcript result results[%d]: %w", i, err)
		}
	}
	return nil
}

func (r CommandRunResult) Validate() error {
	switch r.Kind {
	case "success":
		if r.Failure != nil {
			return errors.New("command.run success result must not include failure")
		}
		if r.ExitCode == nil {
			return errors.New("command.run success result must include exitCode")
		}
	case "failure":
		if r.Failure == nil {
			return errors.New("command.run failure result must include failure")
		}
		if r.ExitCode != nil {
			return errors.New("command.run failure result must not include exitCode")
		}
	default:
		return errors.New("unknown command.run result kind")
	}

	return nil
}
