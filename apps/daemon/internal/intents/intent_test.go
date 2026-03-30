package intents

import (
	"encoding/json"
	"testing"
)

func TestValidateIntentUndo(t *testing.T) {
	t.Parallel()
	err := Intent{
		Kind: IntentKindUndo,
		Undo: &UndoIntent{Scope: UndoScopeLastTranscript},
	}.Validate()
	if err != nil {
		t.Fatalf("expected undo to be valid: %v", err)
	}
}

func TestIntentJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []Intent{
		{
			Kind: IntentKindCommand,
			Command: &CommandIntent{
				Command: "echo",
				Args:    []string{"hi"},
			},
		},
	}
	for _, want := range cases {
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatalf("marshal %+v: %v", want, err)
		}
		var got Intent
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %s: %v", string(data), err)
		}
		if got.Summary() != want.Summary() {
			t.Fatalf("summary: got %q want %q (json=%s)", got.Summary(), want.Summary(), string(data))
		}
	}
}
