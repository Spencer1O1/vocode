package prompt

import (
	"encoding/json"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

func ScopeIntentSystem() string {
	return strings.TrimSpace(`
You are Vocode's scope intent classifier.

You must choose the best scope for the user's voice instruction.
Return exactly one JSON object with this schema:

{
  "scopeKind": "current_function" | "current_file" | "named_symbol" | "clarify",
  "symbolName"?: string,          // required only when scopeKind="named_symbol"
  "clarifyQuestion"?: string      // required only when scopeKind="clarify"
}

Rules:
- Prefer "current_function" for typical "fix/implement/refactor" requests when the cursor is inside a function.
- Use "current_file" when the user explicitly says "in this file" / "whole file" / "entire file".
- Use "named_symbol" when the user explicitly names a function/type/etc to edit (e.g. "in fooBar" / "edit function fooBar").
- Use "clarify" when you cannot safely choose a scope.
- Do not include any extra keys. No markdown fences.
`)
}

func ScopeIntentUserJSON(in agentcontext.ScopeIntentContext) ([]byte, error) {
	type payload struct {
		Instruction      string `json:"instruction"`
		ActiveFile       string `json:"activeFile"`
		CursorSymbol     *struct {
			Name string `json:"name,omitempty"`
			Kind string `json:"kind,omitempty"`
		} `json:"cursorSymbol,omitempty"`
		ActiveFileSymbols []struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"activeFileSymbols,omitempty"`
	}
	var cursor *struct {
		Name string `json:"name,omitempty"`
		Kind string `json:"kind,omitempty"`
	}
	if in.Editor.CursorSymbol != nil {
		cursor = &struct {
			Name string `json:"name,omitempty"`
			Kind string `json:"kind,omitempty"`
		}{Name: in.Editor.CursorSymbol.Name, Kind: in.Editor.CursorSymbol.Kind}
	}
	return json.MarshalIndent(payload{
		Instruction:      strings.TrimSpace(in.Instruction),
		ActiveFile:       in.Editor.ActiveFilePath,
		CursorSymbol:     cursor,
		ActiveFileSymbols: in.ActiveFileSymbols,
	}, "", "  ")
}

