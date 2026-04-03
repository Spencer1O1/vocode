package rootflow

import (
	"context"
	"encoding/json"
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/agent"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// HandleQuestion answers a free-form question using the configured edit model (same client as scoped edits).
func HandleQuestion(deps *RootDeps, _ protocol.VoiceTranscriptParams, text string) (protocol.VoiceTranscriptCompletion, string) {
	q := strings.TrimSpace(text)
	if q == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "question: empty transcript"
	}
	if deps == nil || deps.Editor == nil || deps.Editor.EditModel == nil {
		msg := "No model is configured. Set VOCODE_AGENT_PROVIDER and the required API keys to answer questions by voice."
		return protocol.VoiceTranscriptCompletion{
			Success:       true,
			UiDisposition: "hidden",
			Summary:       msg,
			Question:      &protocol.VoiceTranscriptQuestionAnswer{AnswerText: msg},
		}, ""
	}

	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"answerText"},
		"properties": map[string]any{
			"answerText": map[string]any{"type": "string"},
		},
	}
	sys := strings.TrimSpace(`
You are Vocode's voice assistant inside an IDE. Answer the user's question briefly and clearly.
If you lack repository or editor context, say what you would need. No markdown code fences in the answer text.
Respond with one JSON object only: {"answerText":"..."}. No other keys.
`)
	out, err := deps.Editor.EditModel.Call(context.Background(), agent.CompletionRequest{
		System:     sys,
		User:       q,
		JSONSchema: schema,
	})
	if err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "question model: " + err.Error()
	}
	var parsed struct {
		AnswerText string `json:"answerText"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		return protocol.VoiceTranscriptCompletion{Success: false}, "question: bad model response"
	}
	ans := strings.TrimSpace(parsed.AnswerText)
	if ans == "" {
		return protocol.VoiceTranscriptCompletion{Success: false}, "question: empty model answer"
	}
	return protocol.VoiceTranscriptCompletion{
		Success:       true,
		UiDisposition: "hidden",
		Summary:       ans,
		Question:      &protocol.VoiceTranscriptQuestionAnswer{AnswerText: ans},
	}, ""
}
