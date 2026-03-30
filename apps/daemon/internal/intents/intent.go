package intents

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IntentKind is the executable intent discriminator (host directives only).
type IntentKind string

const (
	IntentKindEdit     IntentKind = "edit"
	IntentKindCommand  IntentKind = "command"
	IntentKindNavigate IntentKind = "navigate"
	IntentKindUndo     IntentKind = "undo"
)

// Intent is a single executable step: maps to at most one protocol directive for the host.
// Turn-only outcomes (irrelevant, done, request_context) live on the agent turn result, not here.
type Intent struct {
	Kind     IntentKind        `json:"kind"`
	Edit     *EditIntent       `json:"edit,omitempty"`
	Command  *CommandIntent    `json:"command,omitempty"`
	Navigate *NavigationIntent `json:"navigate,omitempty"`
	Undo     *UndoIntent       `json:"undo,omitempty"`
}

// Validate checks kind and payload constraints.
func (i Intent) Validate() error {
	return validateIntent(i)
}

// Summary returns a short label (e.g. "edit", "command").
func (i Intent) Summary() string {
	return string(i.Kind)
}

func validateIntent(e Intent) error {
	switch e.Kind {
	case IntentKindEdit:
		if e.Edit == nil {
			return fmt.Errorf("intent: kind %q requires edit", e.Kind)
		}
		return ValidateEditIntent(*e.Edit)
	case IntentKindCommand:
		if e.Command == nil {
			return fmt.Errorf("intent: kind %q requires command", e.Kind)
		}
		if strings.TrimSpace(e.Command.Command) == "" {
			return fmt.Errorf("intent: command.command is empty")
		}
		return nil
	case IntentKindNavigate:
		if e.Navigate == nil {
			return fmt.Errorf("intent: kind %q requires navigate", e.Kind)
		}
		return ValidateNavigationIntent(*e.Navigate)
	case IntentKindUndo:
		if e.Undo == nil {
			return fmt.Errorf("intent: kind %q requires undo", e.Kind)
		}
		return ValidateUndoIntent(*e.Undo)
	default:
		return fmt.Errorf("intent: unknown kind %q", e.Kind)
	}
}

// UnmarshalJSON decodes the flat wire shape (top-level kind + payload).
func (i *Intent) UnmarshalJSON(data []byte) error {
	*i = Intent{}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("intent: %w", err)
	}
	kindRaw, ok := raw["kind"]
	if !ok {
		return fmt.Errorf("intent: missing kind")
	}
	var kind string
	if err := json.Unmarshal(kindRaw, &kind); err != nil {
		return fmt.Errorf("intent: kind: %w", err)
	}
	switch kind {
	case string(IntentKindEdit):
		ex := Intent{Kind: IntentKindEdit}
		ex.Edit = new(EditIntent)
		if err := json.Unmarshal(raw["edit"], ex.Edit); err != nil {
			return fmt.Errorf("intent: edit: %w", err)
		}
		*i = ex
	case string(IntentKindCommand):
		ex := Intent{Kind: IntentKindCommand}
		ex.Command = new(CommandIntent)
		if err := json.Unmarshal(raw["command"], ex.Command); err != nil {
			return fmt.Errorf("intent: command: %w", err)
		}
		*i = ex
	case string(IntentKindNavigate):
		ex := Intent{Kind: IntentKindNavigate}
		ex.Navigate = new(NavigationIntent)
		if err := json.Unmarshal(raw["navigate"], ex.Navigate); err != nil {
			return fmt.Errorf("intent: navigate: %w", err)
		}
		*i = ex
	case string(IntentKindUndo):
		ex := Intent{Kind: IntentKindUndo}
		ex.Undo = new(UndoIntent)
		if err := json.Unmarshal(raw["undo"], ex.Undo); err != nil {
			return fmt.Errorf("intent: undo: %w", err)
		}
		*i = ex
	default:
		return fmt.Errorf("intent: unknown kind %q", kind)
	}
	if err := validateIntent(*i); err != nil {
		return err
	}
	return nil
}

// MarshalJSON encodes the flat wire shape (top-level kind + payload).
func (i Intent) MarshalJSON() ([]byte, error) {
	if err := validateIntent(i); err != nil {
		return nil, err
	}
	out := struct {
		Kind     string            `json:"kind"`
		Edit     *EditIntent       `json:"edit,omitempty"`
		Command  *CommandIntent    `json:"command,omitempty"`
		Navigate *NavigationIntent `json:"navigate,omitempty"`
		Undo     *UndoIntent       `json:"undo,omitempty"`
	}{Kind: string(i.Kind)}
	switch i.Kind {
	case IntentKindEdit:
		out.Edit = i.Edit
	case IntentKindCommand:
		out.Command = i.Command
	case IntentKindNavigate:
		out.Navigate = i.Navigate
	case IntentKindUndo:
		out.Undo = i.Undo
	default:
		return nil, fmt.Errorf("intent: marshal unsupported kind %q", i.Kind)
	}
	return json.Marshal(out)
}
