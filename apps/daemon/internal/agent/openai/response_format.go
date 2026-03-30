package openai

import (
	"encoding/json"
)

// chatResponseFormat picks OpenAI Chat Completions response_format.
// Always uses json_schema with strict=true so the model must return JSON matching the
// turn envelope (kind + optional fields); intent payloads stay validated by turnjson.ParseTurn.
func chatResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &namedJSONSchema{
			Name:   "vocode_turn",
			Strict: true,
			Schema: turnEnvelopeJSONSchema(),
		},
	}
}

type responseFormat struct {
	Type       string           `json:"type"`
	JSONSchema *namedJSONSchema `json:"json_schema,omitempty"`
}

type namedJSONSchema struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict"`
	Schema any    `json:"schema"`
}

func turnEnvelopeJSONSchema() map[string]any {
	// Non-strict: additionalProperties allowed so intent objects stay flexible.
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"kind": map[string]any{
				"type": "string",
				"enum": []string{
					"irrelevant",
					"done",
					"request_context",
					"intents",
				},
			},
			"reason":         map[string]any{"type": "string"},
			"summary":        map[string]any{"type": "string"},
			"requestContext": map[string]any{"type": "object"},
			"intents": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
				},
			},
		},
		"required":             []string{"kind"},
		"additionalProperties": true,
	}
}

// marshalChatResponseFormatJSON exists for tests (stable shape without building a full request).
func marshalChatResponseFormatJSON() ([]byte, error) {
	return json.Marshal(chatResponseFormat())
}
