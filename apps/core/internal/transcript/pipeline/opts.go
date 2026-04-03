package pipeline

import "vocoding.net/vocode/v2/apps/core/internal/flows"

// Opts optionally supplies a route already computed by FlowRouter (classify-then-queue fast path).
type Opts struct {
	HasPreclassified             bool
	PreclassifiedFlow            flows.ID
	PreclassifiedRoute           string
	PreclassifiedSearchQuery     string
	PreclassifiedSearchSymbolKind string
}
