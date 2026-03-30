package prompt

import (
	"encoding/json"
	"testing"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

func TestUserJSONMinimal(t *testing.T) {
	t.Parallel()
	in := agentcontext.TurnContext{
		TranscriptText: "rename foo",
		Editor: agentcontext.EditorSnapshot{
			WorkspaceRoot:  "/ws",
			ActiveFilePath: "/ws/a.go",
		},
	}
	b, err := UserJSON(in)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["transcript"] != "rename foo" {
		t.Fatalf("%s", string(b))
	}
}
