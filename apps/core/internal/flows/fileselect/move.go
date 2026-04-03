package fileselectflow

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/search"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	"vocoding.net/vocode/v2/apps/core/internal/workspace"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// HandleMove moves the focused file or folder to another directory or full path (move_path).
func HandleMove(deps *SelectFileDeps, params protocol.VoiceTranscriptParams, vs *session.VoiceSession, text string) (protocol.VoiceTranscriptCompletion, string) {
	from := strings.TrimSpace(vs.FileSelectionFocus)
	if from == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "move: no file or folder selected"
	}
	if isOpenedWorkspaceRoot(params, from) {
		return protocol.VoiceTranscriptCompletion{Success: false}, "move: cannot move the workspace folder"
	}
	text = strings.TrimSpace(text)
	var dest string
	if deps != nil && deps.Editor != nil && deps.Editor.EditModel != nil {
		var err error
		dest, err = extractMoveDestination(context.Background(), deps.Editor.EditModel, params, from, text)
		if err != nil || strings.TrimSpace(dest) == "" {
			return protocol.VoiceTranscriptCompletion{Success: false},
				"move: could not infer destination from speech — check the model or workspace path"
		}
	} else {
		var ok bool
		dest, ok = parseMoveToDestination(text)
		if !ok {
			return protocol.VoiceTranscriptCompletion{Success: false},
				`move: configure an agent model for natural-language destinations, or say "move … to exact/relative/path"`
		}
	}
	to, errMsg := resolveMoveTarget(params, from, dest)
	if errMsg != "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, errMsg
	}
	if strings.EqualFold(filepath.Clean(from), filepath.Clean(to)) {
		return protocol.VoiceTranscriptCompletion{Success: false}, "move: path unchanged"
	}
	if deps == nil || deps.HostApply == nil || deps.NewBatchID == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply client not configured"
	}

	batchID := deps.NewBatchID()
	dir := []protocol.VoiceTranscriptDirective{
		{Kind: "move_path", MovePathDirective: &protocol.MovePathDirective{From: from, To: to}},
	}
	pending := &session.DirectiveApplyBatch{ID: batchID, NumDirectives: len(dir)}
	vs.PendingDirectiveApply = pending
	hostRes, err := deps.HostApply.ApplyDirectives(protocol.HostApplyParams{
		ApplyBatchId: batchID,
		ActiveFile:   params.ActiveFile,
		Directives:   dir,
	})
	if err != nil {
		vs.PendingDirectiveApply = nil
		return protocol.VoiceTranscriptCompletion{Success: false}, "host.applyDirectives failed: " + err.Error()
	}
	if err := pending.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
		vs.PendingDirectiveApply = nil
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply failed: " + err.Error()
	}
	vs.PendingDirectiveApply = nil
	updateFileSelectionPaths(vs, from, to)
	syncFileSelectionIndexToFocus(vs)

	st, statErr := os.Stat(to)
	isDir := statErr == nil && st.IsDir()
	if !isDir {
		openDirs := searchapply.OpenFirstFileDirectivesForPath(to)
		ob := deps.NewBatchID()
		p2 := &session.DirectiveApplyBatch{ID: ob, NumDirectives: len(openDirs)}
		vs.PendingDirectiveApply = p2
		hostRes2, err := deps.HostApply.ApplyDirectives(protocol.HostApplyParams{
			ApplyBatchId: ob,
			ActiveFile:   params.ActiveFile,
			Directives:   openDirs,
		})
		if err != nil {
			vs.PendingDirectiveApply = nil
			return protocol.VoiceTranscriptCompletion{
				Success:       true,
				Summary:       "moved to " + filepath.Base(to) + " (open failed: " + err.Error() + ")",
				UiDisposition: "browse",
				FileSelection: voiceFileSelectionFromSession(vs),
			}, ""
		}
		_ = p2.ConsumeHostApplyReport(ob, hostRes2.Items)
		vs.PendingDirectiveApply = nil
	}

	comp := protocol.VoiceTranscriptCompletion{
		Success:       true,
		Summary:       "moved to " + filepath.Base(to),
		UiDisposition: "browse",
		FileSelection: voiceFileSelectionFromSession(vs),
	}
	return comp, ""
}

func parseMoveToDestination(text string) (dest string, ok bool) {
	t := strings.ToLower(strings.TrimSpace(text))
	if !strings.Contains(t, "move") {
		return "", false
	}
	lt := strings.ToLower(text)
	idx := strings.LastIndex(lt, " to ")
	if idx < 0 {
		return "", false
	}
	dest = strings.TrimSpace(text[idx+4:])
	dest = strings.Trim(dest, `"'`)
	return dest, dest != ""
}

// normalizeMoveDestinationRel maps a workspace-relative destination before join/resolve.
// If the first path segment names the workspace folder (same as filepath.Base(workspaceRoot)),
// that segment means workspace root — strip it. A destination that is only that name becomes ".".
func normalizeMoveDestinationRel(root, dest string) string {
	dest = strings.TrimSpace(dest)
	if dest == "" || filepath.IsAbs(dest) || dest == "." {
		return dest
	}
	root = filepath.Clean(strings.TrimSpace(root))
	base := filepath.Base(root)
	if base == "" || base == "." {
		return filepath.FromSlash(filepath.ToSlash(dest))
	}
	slash := filepath.ToSlash(dest)
	for {
		segs := moveDestSegments(slash)
		if len(segs) == 0 {
			return "."
		}
		if segs[0] == ".." {
			return filepath.FromSlash(slash)
		}
		if !segmentNamesWorkspaceRoot(segs[0], base) {
			break
		}
		segs = segs[1:]
		if len(segs) == 0 {
			return "."
		}
		slash = strings.Join(segs, "/")
	}
	return filepath.FromSlash(slash)
}

func moveDestSegments(toSlashPath string) []string {
	var out []string
	for _, p := range strings.Split(toSlashPath, "/") {
		p = strings.TrimSpace(p)
		if p == "" || p == "." {
			continue
		}
		out = append(out, p)
	}
	return out
}

func segmentNamesWorkspaceRoot(segment, workspaceFolderBase string) bool {
	if segment == "" || workspaceFolderBase == "" {
		return false
	}
	if strings.EqualFold(segment, workspaceFolderBase) {
		return true
	}
	return search.NormalizePathTokenForMatch(segment) == search.NormalizePathTokenForMatch(workspaceFolderBase)
}

func moveTargetFromFullDest(fullDest, base string) string {
	if st, err := os.Stat(fullDest); err == nil && st.IsDir() {
		return filepath.Join(fullDest, base)
	}
	if st, err := os.Stat(fullDest); err == nil && !st.IsDir() {
		return fullDest
	}
	if filepath.Ext(fullDest) != "" {
		return fullDest
	}
	return filepath.Join(fullDest, base)
}

func resolveMoveTarget(params protocol.VoiceTranscriptParams, fromPath, dest string) (to string, errMsg string) {
	dest = strings.TrimSpace(dest)
	dest = strings.Trim(dest, `"'`)
	if dest == "" {
		return "", "move: empty destination path"
	}
	fromPath = filepath.Clean(strings.TrimSpace(fromPath))
	base := filepath.Base(fromPath)

	var fullDest string
	if filepath.IsAbs(dest) {
		fullDest = filepath.Clean(dest)
	} else {
		root := strings.TrimSpace(params.WorkspaceRoot)
		if root == "" {
			root = strings.TrimSpace(params.FocusedWorkspacePath)
		}
		if root == "" {
			root = workspace.EffectiveWorkspaceRoot(params.WorkspaceRoot, params.ActiveFile)
		}
		root = filepath.Clean(strings.TrimSpace(root))
		if root == "" {
			return "", "move: workspace root or absolute destination required"
		}
		relDest := normalizeMoveDestinationRel(root, dest)
		if resolved, ok := search.ResolveWorkspaceRelativePath(root, relDest); ok {
			fullDest = resolved
		} else {
			fullDest = filepath.Clean(filepath.Join(root, filepath.FromSlash(relDest)))
		}
	}

	to = moveTargetFromFullDest(fullDest, base)
	return to, ""
}

func updateFileSelectionPaths(vs *session.VoiceSession, from, to string) {
	if vs == nil {
		return
	}
	fromC := filepath.Clean(strings.TrimSpace(from))
	toC := filepath.Clean(strings.TrimSpace(to))
	if fromC == "" || toC == "" {
		return
	}
	if strings.EqualFold(strings.TrimSpace(vs.FileSelectionFocus), fromC) {
		vs.FileSelectionFocus = toC
	}
	for i, p := range vs.FileSelectionPaths {
		if strings.EqualFold(filepath.Clean(strings.TrimSpace(p)), fromC) {
			vs.FileSelectionPaths[i] = toC
		}
	}
}
