package dispatch

import (
	"testing"

	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/command"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func TestDirectiveEmpty(t *testing.T) {
	t.Parallel()
	var d Directive
	if !d.IsEmpty() {
		t.Fatal("expected empty")
	}
	if err := d.Validate(); err == nil {
		t.Fatal("expected validate error on empty")
	}
}

func TestDirectiveEditRoundTrip(t *testing.T) {
	t.Parallel()
	ed := &protocol.EditDirective{Kind: "noop"}
	d := directiveEdit(ed)
	if d.IsEmpty() {
		t.Fatal("not empty")
	}
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
	if d.Kind != DirectiveKindEdit || d.EditDirective != ed {
		t.Fatalf("%+v", d)
	}
}

func TestDirectiveFromCommandDispatch(t *testing.T) {
	t.Parallel()
	res, err := command.Dispatch(intents.CommandIntent{Command: "echo", Args: []string{"ok"}})
	if err != nil {
		t.Fatal(err)
	}
	d := directiveCommand(&res)
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDirectiveInvariantExtraPointer(t *testing.T) {
	t.Parallel()
	d := directiveEdit(&protocol.EditDirective{Kind: "noop"})
	d.CommandDirective = &protocol.CommandDirective{}
	if err := d.Validate(); err == nil {
		t.Fatal("expected invariant error")
	}
}
