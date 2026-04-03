package fileselectflow

import (
	"os"
	"path/filepath"
	"testing"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func TestNormalizeMoveDestinationRel(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Evade")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if g := normalizeMoveDestinationRel(root, "Evade/game.js"); g != filepath.FromSlash("game.js") {
		t.Fatalf("got %q want game.js", g)
	}
	if g := normalizeMoveDestinationRel(root, "evade"); g != "." {
		t.Fatalf("root by name: got %q want .", g)
	}
	if g := normalizeMoveDestinationRel(root, "Res/game.js"); g != filepath.FromSlash("Res/game.js") {
		t.Fatalf("got %q", g)
	}
}

func TestResolveMoveTarget_workspaceRootBasename(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Evade")
	if err := os.MkdirAll(filepath.Join(root, "Res"), 0o755); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(root, "Res", "game.js")
	if err := os.WriteFile(from, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	params := protocol.VoiceTranscriptParams{WorkspaceRoot: root}
	to, errMsg := resolveMoveTarget(params, from, "Evade")
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	want := filepath.Join(root, "game.js")
	if filepath.Clean(to) != filepath.Clean(want) {
		t.Fatalf("to %q want %q", to, want)
	}
}

func TestResolveMoveTarget_stripLeadingWorkspaceName(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Evade")
	if err := os.MkdirAll(filepath.Join(root, "Res"), 0o755); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(root, "Res", "game.js")
	if err := os.WriteFile(from, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	params := protocol.VoiceTranscriptParams{WorkspaceRoot: root}
	to, errMsg := resolveMoveTarget(params, from, "Evade/game.js")
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	want := filepath.Join(root, "game.js")
	if filepath.Clean(to) != filepath.Clean(want) {
		t.Fatalf("to %q want %q", to, want)
	}
}

func TestResolveMoveTarget_dotMeansWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "Res"), 0o755); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(root, "Res", "game.js")
	if err := os.WriteFile(from, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	params := protocol.VoiceTranscriptParams{WorkspaceRoot: root}
	to, errMsg := resolveMoveTarget(params, from, ".")
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	want := filepath.Join(root, "game.js")
	if filepath.Clean(to) != filepath.Clean(want) {
		t.Fatalf("to %q want %q", to, want)
	}
}
