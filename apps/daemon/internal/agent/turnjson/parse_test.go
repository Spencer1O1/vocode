package turnjson

import (
	"testing"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
)

func TestParseTurnIrrelevant(t *testing.T) {
	t.Parallel()
	tr, err := ParseTurn([]byte(`{"kind":"irrelevant","reason":"not coding"}`))
	if err != nil {
		t.Fatal(err)
	}
	if tr.Kind != agent.TurnIrrelevant || tr.IrrelevantReason != "not coding" {
		t.Fatalf("%+v", tr)
	}
}

func TestParseTurnIntents(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"kind":"intents","intents":[{"kind":"navigate","navigate":{"kind":"open_file","openFile":{"path":"x.go"}}}]}`)
	tr, err := ParseTurn(raw)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Kind != agent.TurnIntents || len(tr.Intents) != 1 {
		t.Fatalf("%+v", tr)
	}
}

func TestParseTurnGatherContext(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"kind":"request_context","requestContext":{"kind":"request_symbols","query":"foo","maxResult":5}}`)
	tr, err := ParseTurn(raw)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Kind != agent.TurnGatherContext || tr.GatherContext == nil {
		t.Fatalf("%+v", tr)
	}
}

func TestParseTurnWithFence(t *testing.T) {
	t.Parallel()
	raw := "```json\n{\"kind\":\"done\",\"summary\":\"ok\"}\n```"
	tr, err := ParseTurn([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if tr.Kind != agent.TurnFinish || tr.FinishSummary != "ok" {
		t.Fatalf("%+v", tr)
	}
}
