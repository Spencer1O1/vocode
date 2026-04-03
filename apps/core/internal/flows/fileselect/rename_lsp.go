package fileselectflow

import (
	"path/filepath"
	"regexp"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func isLikelyIdentifier(s string) bool {
	ok, err := regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`, s)
	return err == nil && ok
}

func findFirstIdentifierOccurrence(text, ident string) (line, char int, ok bool) {
	if ident == "" {
		return 0, 0, false
	}
	re, err := regexp.Compile(`\b` + regexp.QuoteMeta(ident) + `\b`)
	if err != nil {
		return 0, 0, false
	}
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		loc := re.FindStringIndex(ln)
		if loc != nil {
			return i, loc[0], true
		}
	}
	return 0, 0, false
}

// tryLSPRenameForFile runs vscode.executeDocumentRenameProvider at the first occurrence of oldIdent,
// so references update before the file is moved on disk. Returns true if the host reports success.
func tryLSPRenameForFile(
	deps *SelectFileDeps,
	params protocol.VoiceTranscriptParams,
	vs *session.VoiceSession,
	fromPath, oldIdent, newIdent string,
) bool {
	if deps == nil || deps.HostApply == nil || deps.NewBatchID == nil {
		return false
	}
	if deps.Editor == nil || deps.Editor.ExtensionHost == nil {
		return false
	}
	if oldIdent == newIdent || oldIdent == "" || newIdent == "" {
		return false
	}
	if !isLikelyIdentifier(oldIdent) || !isLikelyIdentifier(newIdent) {
		return false
	}
	body, err := deps.Editor.ExtensionHost.ReadHostFile(fromPath)
	if err != nil {
		return false
	}
	line, char, ok := findFirstIdentifierOccurrence(body, oldIdent)
	if !ok {
		return false
	}

	batchID := deps.NewBatchID()
	dir := []protocol.VoiceTranscriptDirective{
		{
			Kind: "rename",
			RenameDirective: &struct {
				Path     string `json:"path"`
				Position struct {
					Line      int64 `json:"line"`
					Character int64 `json:"character"`
				} `json:"position"`
				NewName string `json:"newName"`
			}{
				Path: fromPath,
				Position: struct {
					Line      int64 `json:"line"`
					Character int64 `json:"character"`
				}{
					Line:      int64(line),
					Character: int64(char),
				},
				NewName: newIdent,
			},
		},
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
		return false
	}
	if err := pending.ConsumeHostApplyReport(batchID, hostRes.Items); err != nil {
		if vs != nil {
			vs.PendingDirectiveApply = nil
		}
		return false
	}
	if vs != nil {
		vs.PendingDirectiveApply = nil
	}
	return true
}

func stemForLSP(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func pickLSPNewName(aiName, newBase string) (string, bool) {
	if isLikelyIdentifier(aiName) {
		return aiName, true
	}
	ns := stemForLSP(newBase)
	if ns != "" && isLikelyIdentifier(ns) {
		return ns, true
	}
	return "", false
}
