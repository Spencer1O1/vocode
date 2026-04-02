package agent

import "context"

// ModelClient performs flow-scoped transcript classification.
type ModelClient interface {
	ClassifyFlow(ctx context.Context, in ClassifierContext) (ClassifierResult, error)
}
