package globalflow

import (
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// TryHandleFileSelectSearch runs file-path search using the classifier-provided file/folder basename.
// file_select never falls back to workspace symbol/content search: that route is for paths on disk only.
// host is the flow that dispatched file_select (Root, SelectFile, or WorkspaceSelect) for preserve behavior.
func TryHandleFileSelectSearch(
	deps *RouteDeps,
	params protocol.VoiceTranscriptParams,
	vs *session.VoiceSession,
	searchQuery string,
	searchSymbolKind string,
	host flows.ID,
) (protocol.VoiceTranscriptCompletion, string, bool) {
	q := strings.TrimSpace(searchQuery)
	if q == "" {
		return protocol.VoiceTranscriptCompletion{}, "", false
	}
	if deps == nil || deps.Search == nil {
		return protocol.VoiceTranscriptCompletion{}, "", false
	}
	if res, hit, reason := deps.Search.FileSearchFromQuery(params, q, vs); hit {
		if strings.TrimSpace(reason) != "" {
			return protocol.VoiceTranscriptCompletion{Success: false}, reason, true
		}
		if res.FileSelection != nil && res.FileSelection.NoHits {
			if host == flows.SelectFile && (len(vs.FileSelectionPaths) > 0 || strings.TrimSpace(vs.FileSelectionFocus) != "") {
				c := selectFileSearchMiss(host, vs)
				c.Summary = fmt.Sprintf("No file path matches for %q — showing your previous selection.", q)
				return c, "", true
			}
		}
		return res, "", true
	}
	return protocol.VoiceTranscriptCompletion{}, "", false
}

// HandleFileSelect handles the global "file_select" route for a sub-flow host or root.
func HandleFileSelect(
	deps *RouteDeps,
	params protocol.VoiceTranscriptParams,
	vs *session.VoiceSession,
	host flows.ID,
	searchQuery string,
	searchSymbolKind string,
) (protocol.VoiceTranscriptCompletion, string) {
	if res, fail, ok := TryHandleFileSelectSearch(deps, params, vs, searchQuery, searchSymbolKind, host); ok {
		return res, fail
	}
	return selectFileSearchMiss(host, vs), ""
}

func selectFileSearchMiss(host flows.ID, vs *session.VoiceSession) protocol.VoiceTranscriptCompletion {
	switch host {
	case flows.Root:
		return protocol.VoiceTranscriptCompletion{
			Success:       true,
			UiDisposition: "skipped",
		}
	case flows.SelectFile:
		c := protocol.VoiceTranscriptCompletion{
			Success:       true,
			Summary:       "file focus updated",
			UiDisposition: "browse",
		}
		if len(vs.FileSelectionPaths) > 0 {
			c.FileSelection = searchapply.FileSearchStateFromPathsWithDir(vs.FileSelectionPaths, vs.FileSelectionIsDir, vs.FileSelectionIndex)
		} else if strings.TrimSpace(vs.FileSelectionFocus) != "" {
			c.FileSelection = searchapply.FileSearchStateFromSinglePath(vs.FileSelectionFocus)
		}
		return c
	default: // flows.WorkspaceSelect
		c := protocol.VoiceTranscriptCompletion{
			Success:       true,
			Summary:       "No file name was provided — say a file or folder name to find.",
			UiDisposition: "hidden",
		}
		if ws := WorkspaceSearchStateFromSession(vs); ws != nil {
			c.Search = ws
			c.Summary = "Keeping current search results; say a file or folder name to search."
			c.UiDisposition = "browse"
		}
		return c
	}
}
