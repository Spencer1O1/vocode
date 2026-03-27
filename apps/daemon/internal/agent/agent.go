package agent

import (
	"context"

	"vocoding.net/vocode/v2/apps/daemon/internal/actionplan"
)

// Agent is the daemon-side runtime facade around the planner [ModelClient].
type Agent struct {
	model ModelClient
}

// New builds an agent with the given [ModelClient] (stub, OpenAI, Anthropic, tests, …).
func New(model ModelClient) *Agent {
	return &Agent{model: model}
}

// NextAction proxies one iterative planner turn.
func (a *Agent) NextAction(ctx context.Context, in ModelInput) (actionplan.NextAction, error) {
	return a.model.NextAction(ctx, in)
}
