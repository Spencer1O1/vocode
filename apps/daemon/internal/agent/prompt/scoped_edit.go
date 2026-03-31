// Package prompt builds model system/user text for narrow, typed model calls.
package prompt

import (
	"encoding/json"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

func ScopedEditSystem() string {
	return strings.TrimSpace(`
You are Vocode's scoped edit model.

You will receive a JSON object describing:
- the user's instruction
- the active file path
- a daemon-resolved target range in that file
- the exact current text inside that range (targetText)

You MUST respond with exactly one JSON object with this schema:

{"replacementText":"..."}

Rules:
- replacementText must be valid source code for the file and should implement the instruction.
- Only edit what is necessary inside the provided target range.
- Do not include markdown fences or any extra keys.
`)
}

func ScopedEditUserJSON(in agentcontext.ScopedEditContext) ([]byte, error) {
	type payload struct {
		Instruction string               `json:"instruction"`
		ActiveFile  string               `json:"activeFile"`
		Target      agentcontext.ResolvedTarget `json:"target"`
		TargetText  string               `json:"targetText"`
	}
	return json.MarshalIndent(payload{
		Instruction: strings.TrimSpace(in.Instruction),
		ActiveFile:  in.Editor.ActiveFilePath,
		Target:      in.Target,
		TargetText:  in.TargetText,
	}, "", "  ")
}

