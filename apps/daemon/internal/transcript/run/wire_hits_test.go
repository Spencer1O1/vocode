package run

import (
	"testing"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

func TestSyncSelectionStackForHits_emptyHits_popsClarifyOnSelection(t *testing.T) {
	t.Parallel()
	vs := &agentcontext.VoiceSession{
		FlowStack: []agentcontext.FlowFrame{
			{Kind: agentcontext.FlowKindSelection},
			{
				Kind:            agentcontext.FlowKindClarify,
				ClarifyQuestion: "which hit?",
			},
		},
	}
	syncSelectionStackForHits(vs)
	if len(vs.FlowStack) != 0 {
		t.Fatalf("expected empty stack, got %#v", vs.FlowStack)
	}
}

func TestSyncSelectionStackForHits_emptyHits_keepsStandaloneClarify(t *testing.T) {
	t.Parallel()
	vs := &agentcontext.VoiceSession{
		FlowStack: []agentcontext.FlowFrame{
			{
				Kind:            agentcontext.FlowKindClarify,
				ClarifyQuestion: "which file?",
			},
		},
	}
	syncSelectionStackForHits(vs)
	if len(vs.FlowStack) != 1 || vs.FlowStack[0].Kind != agentcontext.FlowKindClarify {
		t.Fatalf("expected lone clarify retained, got %#v", vs.FlowStack)
	}
}

func TestSyncSelectionStackForHits_emptyHits_popsSelectionOnly(t *testing.T) {
	t.Parallel()
	vs := &agentcontext.VoiceSession{
		FlowStack: []agentcontext.FlowFrame{
			{Kind: agentcontext.FlowKindSelection},
		},
	}
	syncSelectionStackForHits(vs)
	if len(vs.FlowStack) != 0 {
		t.Fatalf("expected empty stack, got %#v", vs.FlowStack)
	}
}

func TestSyncSelectionStackForHits_withHits_doesNotStackDuplicateSelection(t *testing.T) {
	t.Parallel()
	vs := &agentcontext.VoiceSession{
		SearchResults: []agentcontext.SearchHit{{Path: "/a.go", Line: 1, Character: 0, Preview: "x"}},
		FlowStack: []agentcontext.FlowFrame{
			{Kind: agentcontext.FlowKindSelection},
		},
	}
	syncSelectionStackForHits(vs)
	if len(vs.FlowStack) != 1 {
		t.Fatalf("expected single selection frame, got len=%d stack=%#v", len(vs.FlowStack), vs.FlowStack)
	}
}
