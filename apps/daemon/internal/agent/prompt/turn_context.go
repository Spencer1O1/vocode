// Package prompt builds model system/user text from [agentcontext.TurnContext]. Daemon-owned (not protocol policy).
package prompt

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

const maxExcerptRunes = 4000
const maxExcerptsInPrompt = 6

// System returns the fixed system instructions for one transcript planner turn.
func System() string {
	return strings.TrimSpace(`
You are Vocode's voice coding planner. The user spoke; you receive IDE context and prior outcomes.
Respond with exactly one JSON object (no markdown fences, no extra text) matching this union:

- {"kind":"irrelevant","reason":"optional short string"} — utterance is not an instruction for the editor (chit-chat, unrelated speech).
- {"kind":"done","summary":"optional string"} — you are finished for this turn; no more host directives.
- {"kind":"request_context","requestContext":{...}} — turn-level only: ask the daemon to enrich context (symbols, file excerpt, usages) before your next completion; not an entry in the intents array.
- {"kind":"intents","intents":[...]} — one or more host intents in order (navigate, edit, command, undo). Each element uses top-level "kind" plus that kind's payload only.

Rules:
- After partial host apply, output only outstanding executables; never repeat items that already succeeded on the host (see attemptHistory with status ok).
- Prefer concise executables; batch navigate + edit + command when the user asked for multiple steps.
- Use only executable kinds and fields the host supports; invalid JSON or unknown kinds will fail validation.
`)
}

// UserJSON renders structured turn input as compact JSON for the user message.
func UserJSON(in agentcontext.TurnContext) ([]byte, error) {
	p := turnPromptPayload{
		Transcript: in.TranscriptText,
		Editor: editorPayload{
			WorkspaceRoot: in.Editor.WorkspaceRoot,
			ActiveFile:    in.Editor.ActiveFilePath,
		},
		AttemptHistory:   attemptHistoryToWire(in),
		Gathered:         gatheredToWire(in.Gathered),
	}
	if in.Editor.CursorSymbol != nil {
		p.Editor.Cursor = &cursorPayload{
			ID:   in.Editor.CursorSymbol.ID,
			Name: in.Editor.CursorSymbol.Name,
			Kind: in.Editor.CursorSymbol.Kind,
		}
	}
	return json.MarshalIndent(p, "", "  ")
}

type turnPromptPayload struct {
	Transcript     string            `json:"transcript"`
	Editor         editorPayload     `json:"editor"`
	AttemptHistory []attemptWire     `json:"attemptHistory,omitempty"`
	Gathered       gatheredWire      `json:"gathered"`
}

type editorPayload struct {
	WorkspaceRoot string         `json:"workspaceRoot"`
	ActiveFile    string         `json:"activeFile"`
	Cursor        *cursorPayload `json:"cursor,omitempty"`
}

type cursorPayload struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type attemptWire struct {
	BatchOrdinal *int            `json:"batchOrdinal,omitempty"`
	IndexInBatch *int            `json:"indexInBatch,omitempty"`
	Status       string         `json:"status"`
	Phase        string         `json:"phase,omitempty"`
	Message      string         `json:"message,omitempty"`
	Intent       json.RawMessage `json:"intent"`
}

type gatheredWire struct {
	Symbols  []symbolWire  `json:"symbols,omitempty"`
	Excerpts []excerptWire `json:"excerpts,omitempty"`
	Notes    []string      `json:"notes,omitempty"`
}

type symbolWire struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type excerptWire struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func attemptHistoryToWire(in agentcontext.TurnContext) []attemptWire {
	// Host apply outcomes (persisted in IntentApplyHistory) already include ok/failed/skipped
	// plus the most relevant per-directive message.
	out := make([]attemptWire, 0, len(in.IntentApplyHistory)+len(in.FailedIntents))

	if len(in.IntentApplyHistory) > 0 {
		for _, h := range in.IntentApplyHistory {
			b, err := json.Marshal(h.Intent)
			if err != nil {
				continue
			}
			bo := h.BatchOrdinal
			i := h.IndexInBatch
			out = append(out, attemptWire{
				BatchOrdinal: &bo,
				IndexInBatch: &i,
				Status:       string(h.Status),
				Message:      h.Message,
				Intent:       b,
			})
		}
	}

	// Dispatch-time failures are not part of IntentApplyHistory yet, so we append them.
	if len(in.FailedIntents) > 0 {
		for _, f := range in.FailedIntents {
			b, err := json.Marshal(f.Intent)
			if err != nil {
				continue
			}
			out = append(out, attemptWire{
				Status:  "failed",
				Phase:   f.Phase,
				Message: f.Reason,
				Intent:  b,
			})
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func gatheredToWire(g agentcontext.Gathered) gatheredWire {
	w := gatheredWire{
		Notes: append([]string(nil), g.Notes...),
	}
	for _, s := range g.Symbols {
		w.Symbols = append(w.Symbols, symbolWire{
			ID:   s.ID,
			Name: s.Name,
			Path: s.Path,
			Kind: s.Kind,
		})
	}
	for i, ex := range g.Excerpts {
		if i >= maxExcerptsInPrompt {
			break
		}
		content := ex.Content
		if utf8.RuneCountInString(content) > maxExcerptRunes {
			runes := []rune(content)
			content = string(runes[:maxExcerptRunes]) + "\n…(truncated)"
		}
		w.Excerpts = append(w.Excerpts, excerptWire{Path: ex.Path, Content: content})
	}
	return w
}
