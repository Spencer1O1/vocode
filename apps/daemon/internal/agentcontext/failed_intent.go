package agentcontext

import "vocoding.net/vocode/v2/apps/daemon/internal/intents"

// FailurePhase identifies where an intent failed on the path to a completed host action.
const (
	PhasePreExecute     = "pre_execute"     // daemon rejected before dispatch (e.g. missing active file for edit)
	PhaseDispatch       = "dispatch"        // dispatch.Handler returned an error for an executable
	PhaseContextFulfill = "context_fulfill" // request_context fulfillment error (not currently wrapped as FailedIntent)
	// PhaseExtension is when the extension reports a directive apply failure via lastBatchApply.
	PhaseExtension = "extension"
)

// FailedIntent records an intent that was rejected (with where and why) so the model can retry coherently.
// It is not a separate intent kind—same payload type as [intents.Intent], plus diagnostics.
type FailedIntent struct {
	Intent intents.Intent
	Phase  string
	Reason string
}
