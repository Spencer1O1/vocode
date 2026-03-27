package agent_test

import (
	"context"
	"testing"

	"vocoding.net/vocode/v2/apps/daemon/internal/actionplan"
	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agent/stub"
)

func TestNextActionStubFlow(t *testing.T) {
	t.Parallel()

	a := agent.New(stub.New())
	in := agent.ModelInput{Transcript: "hello"}

	for i := 0; i < 4; i++ {
		next, err := a.NextAction(context.Background(), in)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if err := actionplan.ValidateNextAction(next); err != nil {
			t.Fatalf("invalid next action: %v", err)
		}
		if next.Kind == actionplan.NextActionKindDone {
			t.Fatal("unexpected done before 4th step")
		}
		in.CompletedActions = append(in.CompletedActions, next)
	}

	final, err := a.NextAction(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if final.Kind != actionplan.NextActionKindDone {
		t.Fatalf("expected done, got %q", final.Kind)
	}
}
