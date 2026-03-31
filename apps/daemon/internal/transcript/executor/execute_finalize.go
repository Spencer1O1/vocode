package executor

import (
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func finalizeExecute(st *agentLoopState) (protocol.VoiceTranscriptResult, []protocol.VoiceTranscriptDirective, agentcontext.Gathered, *agentcontext.DirectiveApplyBatch, bool) {
	result := protocol.VoiceTranscriptResult{
		Success: true,
		Summary: st.transcriptSummary,
	}
	if strings.TrimSpace(st.transcriptOutcome) == "irrelevant" {
		result.TranscriptOutcome = "irrelevant"
	}
	dirs := append([]protocol.VoiceTranscriptDirective(nil), st.directives...)
	var pending *agentcontext.DirectiveApplyBatch
	if len(dirs) > 0 {
		bid, err := newDirectiveApplyBatchID()
		if err != nil {
			return protocol.VoiceTranscriptResult{Success: false}, nil, st.gathered, nil, true
		}
		pending = &agentcontext.DirectiveApplyBatch{
			ID:            bid,
			SourceIntents: append([]intents.Intent(nil), st.batchSourceIntents...),
		}
	}
	if err := result.Validate(); err != nil {
		return protocol.VoiceTranscriptResult{Success: false}, nil, st.gathered, nil, true
	}
	return result, dirs, st.gathered, pending, true
}
