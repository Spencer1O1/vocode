package agentcontext

import "vocoding.net/vocode/v2/apps/daemon/internal/intents"

// TurnContext is everything the agent model sees for one iterative NextIntent call.
type TurnContext struct {
	TranscriptText   string
	SucceededIntents []intents.Intent
	FailedIntents    []FailedIntent
	Editor           EditorSnapshot
	Gathered         Gathered
}
