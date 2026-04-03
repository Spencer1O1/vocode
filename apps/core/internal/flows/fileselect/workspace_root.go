package fileselectflow

import (
	"path/filepath"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/workspace"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// isOpenedWorkspaceRoot reports whether p is the same folder as an opened workspace root we know
// from the transcript params (aligned with VS Code: you cannot delete/rename/move the workspace folder itself).
// Selection may still be the root for create_entry and other non-destructive routes.
//
// Do not use FocusedWorkspacePath here: the host sends the active editor path (and file-selection sync
// uses the same field per schema), so it often equals a selected file path — not a workspace folder root.
func isOpenedWorkspaceRoot(params protocol.VoiceTranscriptParams, p string) bool {
	p = filepath.Clean(strings.TrimSpace(p))
	if p == "" || p == "." {
		return false
	}
	candidates := []string{
		strings.TrimSpace(params.WorkspaceRoot),
		workspace.EffectiveWorkspaceRoot(params.WorkspaceRoot, params.ActiveFile),
	}
	seen := make(map[string]struct{})
	for _, r := range candidates {
		r = filepath.Clean(strings.TrimSpace(r))
		if r == "" {
			continue
		}
		key := strings.ToLower(r)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		if strings.EqualFold(r, p) {
			return true
		}
	}
	return false
}
