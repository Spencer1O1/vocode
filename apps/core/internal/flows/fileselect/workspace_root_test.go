package fileselectflow

import (
	"path/filepath"
	"testing"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func TestIsOpenedWorkspaceRoot(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	params := protocol.VoiceTranscriptParams{
		WorkspaceRoot: root,
	}
	if !isOpenedWorkspaceRoot(params, root) {
		t.Fatal("expected workspace root")
	}
	if !isOpenedWorkspaceRoot(params, root+string(filepath.Separator)) {
		t.Fatal("expected trailing sep normalized")
	}
	sub := filepath.Join(root, "pkg")
	if isOpenedWorkspaceRoot(params, sub) {
		t.Fatal("subfolder is not root")
	}
}

func TestIsOpenedWorkspaceRoot_focusedPathIsActiveFileNotRoot(t *testing.T) {
	root := filepath.Clean(t.TempDir())
	file := filepath.Join(root, "Res", "game.js")
	params := protocol.VoiceTranscriptParams{
		WorkspaceRoot:        root,
		FocusedWorkspacePath: file,
		ActiveFile:           file,
	}
	if isOpenedWorkspaceRoot(params, file) {
		t.Fatal("active/selected file path must not be treated as workspace root")
	}
	if !isOpenedWorkspaceRoot(params, root) {
		t.Fatal("actual workspace folder must still match")
	}
}
