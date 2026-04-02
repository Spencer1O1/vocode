package prompt

import "vocoding.net/vocode/v2/apps/core/internal/flows"

// FlowClassifierJSONSchema is the OpenAI response_format schema for route-only classification.
func FlowClassifierJSONSchema(flow flows.ID) map[string]any {
	routes := flows.SpecFor(flow).RouteIDs()
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"route": map[string]any{
				"type": "string",
				"enum": routes,
			},
		},
		"required":             []string{"route"},
		"additionalProperties": false,
	}
}
