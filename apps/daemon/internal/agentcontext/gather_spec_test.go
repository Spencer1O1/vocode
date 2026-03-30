package agentcontext

import "testing"

func TestGatherContextSpecValidate(t *testing.T) {
	t.Parallel()
	if err := (&GatherContextSpec{Kind: GatherContextKindSymbols}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (&GatherContextSpec{Kind: "nope"}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}
