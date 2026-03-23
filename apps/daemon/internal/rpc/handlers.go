package rpc

import "vocoding.net/vocode/v2/apps/daemon/internal/edits"

type HandlerDefinition struct {
	Method  string
	Handler Handler
}

func BuildHandlers(editService *edits.Service) []HandlerDefinition {
	return []HandlerDefinition{
		{
			Method:  "ping",
			Handler: NewPingHandler(),
		},
		{
			Method:  "edit/apply",
			Handler: NewEditApplyHandler(editService),
		},
	}
}
