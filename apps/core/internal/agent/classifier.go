package agent

import (
	"fmt"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

// ClassifierContext is input for flow-scoped transcript classification.
type ClassifierContext struct {
	Flow        flows.ID
	Instruction string
	Editor      EditorSnapshot

	HitCount    int
	ActiveIndex int

	FocusPath       string
	ListCount       int
	ListActiveIndex int
}

// ClassifierResult is only which route the transcript belongs to.
// Route-specific handlers (heuristic or model) interpret the transcript later.
type ClassifierResult struct {
	Flow  flows.ID
	Route string
}

func (r ClassifierResult) Validate() error {
	if r.Flow != flows.Root && r.Flow != flows.Select && r.Flow != flows.SelectFile {
		return fmt.Errorf("flow classifier: unknown flow %q", r.Flow)
	}
	return flows.ValidateRoute(r.Flow, r.Route)
}
