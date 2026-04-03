package globalflow

import (
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// TryHandleSelectFileSearch runs file-path search using the classifier-provided fragment.
// If path search yields noHits, it tries workspace (symbol/content) search with the same query so
// utterances like "main" still resolve when the classifier wrongly chose select_file.
// host is the flow that dispatched select_file (Root, SelectFile, or WorkspaceSelect) for preserve behavior.
func TryHandleSelectFileSearch(
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
			wr, wfail, wok := TryHandleWorkspaceSelectSearch(deps, params, vs, q, strings.TrimSpace(searchSymbolKind))
			if wok && strings.TrimSpace(wfail) != "" {
				return protocol.VoiceTranscriptCompletion{Success: false}, wfail, true
			}
			if wok && wr.Search != nil && len(wr.Search.Results) > 0 {
				return wr, "", true
			}
			if host == flows.SelectFile && (len(vs.FileSelectionPaths) > 0 || strings.TrimSpace(vs.FileSelectionFocus) != "") {
				c := selectFileSearchMiss(host, vs)
				c.Summary = fmt.Sprintf("no file path matches for %q", q)
				return c, "", true
			}
		}
		return res, "", true
	}
	return protocol.VoiceTranscriptCompletion{}, "", false
}

// HandleSelectFile handles the global "select_file" route for a sub-flow host or root.
func HandleSelectFile(
	deps *RouteDeps,
	params protocol.VoiceTranscriptParams,
	vs *session.VoiceSession,
	host flows.ID,
	searchQuery string,
	searchSymbolKind string,
) (protocol.VoiceTranscriptCompletion, string) {
	if res, fail, ok := TryHandleSelectFileSearch(deps, params, vs, searchQuery, searchSymbolKind, host); ok {
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
		return protocol.VoiceTranscriptCompletion{
			Success:       true,
			Summary:       "core transcript (stub)",
			UiDisposition: "hidden",
		}
	}
}
