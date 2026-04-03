package globalflow

import (
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// WorkspaceSearchStateFromSession builds workspace search UI state from the voice session, or nil if empty.
func WorkspaceSearchStateFromSession(vs *session.VoiceSession) *protocol.VoiceTranscriptWorkspaceSearchState {
	if vs == nil || len(vs.SearchResults) == 0 {
		return nil
	}
	results := make([]protocol.VoiceTranscriptSearchHit, 0, len(vs.SearchResults))
	for _, h := range vs.SearchResults {
		ml := int64(h.Len)
		if ml <= 0 {
			ml = 1
		}
		p := ml
		results = append(results, protocol.VoiceTranscriptSearchHit{
			Path:        h.Path,
			Line:        int64(h.Line),
			Character:   int64(h.Character),
			Preview:     h.Preview,
			MatchLength: &p,
		})
	}
	ai := int64(vs.ActiveSearchIndex)
	return &protocol.VoiceTranscriptWorkspaceSearchState{
		Results:     results,
		ActiveIndex: &ai,
	}
}
