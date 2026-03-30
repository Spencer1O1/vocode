package dispatch

import (
	"fmt"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// DirectiveKind selects which payload field of [Directive] is set (tagged union).
type DirectiveKind string

const (
	DirectiveKindEdit     DirectiveKind = "edit"
	DirectiveKindCommand  DirectiveKind = "command"
	DirectiveKindNavigate DirectiveKind = "navigate"
	DirectiveKindUndo     DirectiveKind = "undo"
)

// Directive is one host-bound directive from [Handler.Handle]: exactly one variant (Kind + matching pointer), or empty (Kind "").
type Directive struct {
	Kind                DirectiveKind
	EditDirective       *protocol.EditDirective
	CommandDirective    *protocol.CommandDirective
	NavigationDirective *protocol.NavigationDirective
	UndoDirective       *protocol.UndoDirective
}

// IsEmpty reports the zero value (no directive produced).
func (d Directive) IsEmpty() bool {
	return d.Kind == ""
}

// Validate checks the tagged-union invariant (Kind matches exactly one non-nil pointer).
func (d Directive) Validate() error {
	if d.Kind == "" {
		return fmt.Errorf("directive: empty")
	}
	switch d.Kind {
	case DirectiveKindEdit:
		if d.EditDirective == nil {
			return fmt.Errorf("directive: kind %q requires EditDirective", d.Kind)
		}
		if d.CommandDirective != nil || d.NavigationDirective != nil || d.UndoDirective != nil {
			return fmt.Errorf("directive: kind %q must not set other pointers", d.Kind)
		}
	case DirectiveKindCommand:
		if d.CommandDirective == nil {
			return fmt.Errorf("directive: kind %q requires CommandDirective", d.Kind)
		}
		if d.EditDirective != nil || d.NavigationDirective != nil || d.UndoDirective != nil {
			return fmt.Errorf("directive: kind %q must not set other pointers", d.Kind)
		}
	case DirectiveKindNavigate:
		if d.NavigationDirective == nil {
			return fmt.Errorf("directive: kind %q requires NavigationDirective", d.Kind)
		}
		if d.EditDirective != nil || d.CommandDirective != nil || d.UndoDirective != nil {
			return fmt.Errorf("directive: kind %q must not set other pointers", d.Kind)
		}
	case DirectiveKindUndo:
		if d.UndoDirective == nil {
			return fmt.Errorf("directive: kind %q requires UndoDirective", d.Kind)
		}
		if d.EditDirective != nil || d.CommandDirective != nil || d.NavigationDirective != nil {
			return fmt.Errorf("directive: kind %q must not set other pointers", d.Kind)
		}
	default:
		return fmt.Errorf("directive: unknown kind %q", d.Kind)
	}
	return nil
}

func directiveEdit(ed *protocol.EditDirective) Directive {
	return Directive{Kind: DirectiveKindEdit, EditDirective: ed}
}

func directiveCommand(cd *protocol.CommandDirective) Directive {
	return Directive{Kind: DirectiveKindCommand, CommandDirective: cd}
}

func directiveNavigate(nd *protocol.NavigationDirective) Directive {
	return Directive{Kind: DirectiveKindNavigate, NavigationDirective: nd}
}

func directiveUndo(ud *protocol.UndoDirective) Directive {
	return Directive{Kind: DirectiveKindUndo, UndoDirective: ud}
}
