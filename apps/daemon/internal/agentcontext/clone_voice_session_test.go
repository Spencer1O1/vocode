package agentcontext

import "testing"

func TestCloneVoiceSession_excerptIsolation(t *testing.T) {
	t.Parallel()
	orig := VoiceSession{
		Gathered: Gathered{
			Excerpts: []FileExcerpt{{Path: "x.go", Content: "alpha"}},
			Notes:    []string{"n1"},
		},
	}
	cl := CloneVoiceSession(orig)
	if len(cl.Gathered.Excerpts) != 1 {
		t.Fatal("expected one excerpt")
	}
	cl.Gathered.Excerpts[0].Content = "beta"
	cl.Gathered.Notes[0] = "changed"
	if orig.Gathered.Excerpts[0].Content != "alpha" {
		t.Fatalf("clone mutated original excerpt: %q", orig.Gathered.Excerpts[0].Content)
	}
	if orig.Gathered.Notes[0] != "n1" {
		t.Fatalf("clone mutated original notes: %q", orig.Gathered.Notes[0])
	}
}

func TestCloneVoiceSession_pendingBatchPointerIsolation(t *testing.T) {
	t.Parallel()
	b := DirectiveApplyBatch{ID: "b1", NumDirectives: 2}
	orig := VoiceSession{PendingDirectiveApply: &b}
	cl := CloneVoiceSession(orig)
	if cl.PendingDirectiveApply == nil || cl.PendingDirectiveApply == orig.PendingDirectiveApply {
		t.Fatal("expected distinct pending batch pointer")
	}
	cl.PendingDirectiveApply.ID = "changed"
	if orig.PendingDirectiveApply.ID != "b1" {
		t.Fatalf("clone mutated original batch id: %q", orig.PendingDirectiveApply.ID)
	}
}
