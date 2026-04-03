package fileselectflow

import (
	"path/filepath"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// syncFileSelectionIndexToFocus aligns FileSelectionIndex with FileSelectionFocus after path updates.
func syncFileSelectionIndexToFocus(vs *session.VoiceSession) {
	if vs == nil || len(vs.FileSelectionPaths) == 0 {
		return
	}
	want := filepath.Clean(strings.TrimSpace(vs.FileSelectionFocus))
	if want == "" {
		return
	}
	for i, p := range vs.FileSelectionPaths {
		if strings.EqualFold(filepath.Clean(strings.TrimSpace(p)), want) {
			vs.FileSelectionIndex = i
			return
		}
	}
}

// voiceFileSelectionFromSession builds protocol file list state for the host sidebar, or nil if empty.
func voiceFileSelectionFromSession(vs *session.VoiceSession) *protocol.VoiceTranscriptFileSearchState {
	if vs == nil || len(vs.FileSelectionPaths) == 0 {
		return nil
	}
	return searchapply.FileSearchStateFromPathsWithDir(
		vs.FileSelectionPaths, vs.FileSelectionIsDir, vs.FileSelectionIndex,
	)
}
