package actionplan

import "testing"

func TestValidateNextActionDone(t *testing.T) {
	t.Parallel()
	if err := ValidateNextAction(NextAction{Kind: NextActionKindDone}); err != nil {
		t.Fatalf("expected done action to be valid: %v", err)
	}
}

func TestValidateNextActionEdit(t *testing.T) {
	t.Parallel()
	a := NextAction{
		Kind: NextActionKindEdit,
		Edit: &EditIntent{
			Kind: EditIntentKindReplace,
			Replace: &ReplaceEditIntent{
				Target: EditTarget{
					Kind:     EditTargetKindSymbolID,
					SymbolID: &SymbolIDTarget{ID: "v1|Zm9vLnRz|1|ZnVuY3Rpb24|YmFy"},
				},
				NewText: "x",
			},
		},
	}
	if err := ValidateNextAction(a); err != nil {
		t.Fatalf("expected no err: %v", err)
	}
}
