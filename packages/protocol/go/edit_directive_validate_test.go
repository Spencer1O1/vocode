package protocol

import "testing"

func TestEditDirectiveValidate(t *testing.T) {
	t.Parallel()

	success := NewEditDirectiveSuccess([]EditAction{{
		Kind: "replace_between_anchors",
		Path: "/tmp/file.ts",
		Anchor: &Anchor{
			Before: "before",
			After:  "after",
		},
		NewText: "updated",
	}})
	if err := success.Validate(); err != nil {
		t.Fatalf("expected success to validate, got %v", err)
	}

	noop := NewEditDirectiveNoop("No change needed.")
	if err := noop.Validate(); err != nil {
		t.Fatalf("expected noop to validate, got %v", err)
	}
}

func TestEditDirectiveValidateRejectsInvalidCombinations(t *testing.T) {
	t.Parallel()

	invalid := []EditDirective{
		{Kind: "success"},
		{
			Kind:    "success",
			Actions: []EditAction{},
			Reason:  "bad edit",
		},
		{Kind: "noop"},
		{Kind: "unknown"},
	}

	for i, candidate := range invalid {
		if err := candidate.Validate(); err == nil {
			t.Fatalf("expected invalid result %d to fail validation", i)
		}
	}
}
