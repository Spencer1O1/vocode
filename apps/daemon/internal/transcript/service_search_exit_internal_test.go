package transcript

import (
	"io"
	"log"
	"testing"
	"time"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agent/stub"
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/transcript/voicesession"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func TestAcceptTranscript_searchControl_exitClearsSession(t *testing.T) {
	t.Helper()
	ag := agent.New(stub.New())
	svc := NewService(ag, log.New(io.Discard, "", 0))
	svc.queue = nil

	key := "session-key-exit-1"
	voicesession.SaveKeyed(svc.sessions, key, agentcontext.VoiceSession{
		SearchResults: []agentcontext.SearchHit{
			{Path: "/a.go", Line: 2, Character: 0, Preview: "one"},
		},
		ActiveSearchIndex: 0,
	})

	res, ok, reason := svc.AcceptTranscript(protocol.VoiceTranscriptParams{
		ContextSessionId: key,
		Text:             "exit",
	})
	if !ok || !res.Success || reason != "" {
		t.Fatalf("got ok=%v success=%v reason=%q res=%+v", ok, res.Success, reason, res)
	}
	if res.TranscriptOutcome != "search_control" {
		t.Fatalf("expected outcome=search_control, got %q", res.TranscriptOutcome)
	}
	if res.UiDisposition != "hidden" {
		t.Fatalf("expected uiDisposition=hidden, got %q", res.UiDisposition)
	}

	loaded := voicesession.Load(svc.sessions, key, time.Hour, nil)
	if len(loaded.SearchResults) != 0 {
		t.Fatalf("expected SearchResults cleared, got %+v", loaded.SearchResults)
	}
}

