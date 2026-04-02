package openai

import (
	"vocoding.net/vocode/v2/apps/core/internal/agent/prompt"
	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

func chatResponseFormatFlowClassifier(flow flows.ID) *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &namedJSONSchema{
			Name:   "vocode_flow_classifier",
			Strict: false,
			Schema: prompt.FlowClassifierJSONSchema(flow),
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
