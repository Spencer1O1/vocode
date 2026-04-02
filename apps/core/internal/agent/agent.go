package agent

import "context"

// Agent is the core-owned pipeline facade.
type Agent struct {
	model ModelClient
}

func New(model ModelClient) *Agent {
	return &Agent{model: model}
}

func (a *Agent) ClassifyFlow(ctx context.Context, in ClassifierContext) (ClassifierResult, error) {
	return a.model.ClassifyFlow(ctx, in)
}
