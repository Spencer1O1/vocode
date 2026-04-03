package fileselectflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	workspaceselectflow "vocoding.net/vocode/v2/apps/core/internal/flows/workspaceselect"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// HandleRename renames the focused file or folder: optional LSP symbol rename (references) for files,
// then move_path on disk. The new name comes from "rename … to …" or, when a model is configured, from an AI extract.
func HandleRename(deps *SelectFileDeps, params protocol.VoiceTranscriptParams, vs *session.VoiceSession, text string) (protocol.VoiceTranscriptCompletion, string) {
	from := strings.TrimSpace(vs.FileSelectionFocus)
	if from == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "rename: no file or folder selected"
	}
	if isOpenedWorkspaceRoot(params, from) {
		return protocol.VoiceTranscriptCompletion{Success: false}, "rename: cannot rename the workspace folder"
	}

	text = strings.TrimSpace(text)
	var newBase string
	var aiLspName string
	if n, ok := workspaceselectflow.ParseSpokenRenameNewName(text); ok {
		newBase = n
	} else if deps != nil && deps.Editor != nil && deps.Editor.EditModel != nil {
		var err error
		newBase, aiLspName, err = extractRenameBasename(context.Background(), deps.Editor.EditModel, params, from, text)
		if err != nil || strings.TrimSpace(newBase) == "" {
			return protocol.VoiceTranscriptCompletion{Success: false},
				"rename: could not infer the new name — say \"rename … to <name>\" or check the model"
		}
	} else {
		return protocol.VoiceTranscriptCompletion{Success: false},
			`rename: use "rename … to <newName>" or configure a model for free-form rename`
	}

	newBase = sanitizeNewFileName(newBase)
	if newBase == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "rename: invalid new name"
	}
	to := filepath.Join(filepath.Dir(from), newBase)
	if strings.EqualFold(filepath.Clean(from), filepath.Clean(to)) {
		return protocol.VoiceTranscriptCompletion{Success: false}, "rename: name unchanged"
	}
	if deps == nil || deps.HostApply == nil || deps.NewBatchID == nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply client not configured"
	}

	st, statErr := os.Stat(from)
	isDir := statErr == nil && st.IsDir()
	if !isDir {
		oldStem := stemForLSP(from)
		if newSym, ok := pickLSPNewName(aiLspName, newBase); ok && oldStem != "" && oldStem != newSym {
			tryLSPRenameForFile(deps, params, vs, from, oldStem, newSym)
		}
	}

	batchID := deps.NewBatchID()
	dir := []protocol.VoiceTranscriptDirective{
		{Kind: "move_path", MovePathDirective: &protocol.MovePathDirective{From: from, To: to}},
	}
	pending := &session.DirectiveApplyBatch{ID: batchID, NumDirectives: len(dir)}
	if vs != nil {
		vs.PendingDirectiveApply = pending
	}
	hostRes, err := deps.HostApply.ApplyDirectives(protocol.HostApplyParams{
		ApplyBatchId: batchID,
		ActiveFile:   params.ActiveFile,
		Directives:   dir,
	})
	if err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, "host.applyDirectives failed: " + err.Error()
	}
	if err := pending.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return protocol.VoiceTranscriptCompletion{Success: false}, "host apply failed: " + err.Error()
	}
	if vs != nil {
		vs.PendingDirectiveApply = nil
	}
	updateFileSelectionPaths(vs, from, to)
	syncFileSelectionIndexToFocus(vs)

	if !isDir {
		openDirs := searchapply.OpenFirstFileDirectivesForPath(to)
		ob := deps.NewBatchID()
		p2 := &session.DirectiveApplyBatch{ID: ob, NumDirectives: len(openDirs)}
		if vs != nil {
			vs.PendingDirectiveApply = p2
		}
		hostRes2, err := deps.HostApply.ApplyDirectives(protocol.HostApplyParams{
			ApplyBatchId: ob,
			ActiveFile:   params.ActiveFile,
			Directives:   openDirs,
		})
		if err != nil {
			if vs != nil {
				vs.PendingDirectiveApply = nil
			}
			return protocol.VoiceTranscriptCompletion{
				Success:       true,
				Summary:       fmt.Sprintf("renamed to %s (open failed: %v)", newBase, err),
				UiDisposition: "browse",
				FileSelection: voiceFileSelectionFromSession(vs),
			}, ""
		}
		_ = p2.ConsumeHostApplyReport(ob, hostRes2.Items)
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
	}

	return protocol.VoiceTranscriptCompletion{
		Success:       true,
		Summary:       "renamed to " + newBase,
		UiDisposition: "browse",
		FileSelection: voiceFileSelectionFromSession(vs),
	}, ""
}
