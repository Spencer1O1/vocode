package executor

import (
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// loopAdvance controls whether Execute continues after one NextTurn + dispatch cycle.
type loopAdvance byte

const (
	advanceContinue loopAdvance = iota
	advanceBreakLoop
	// advanceBatchIntentDone: executable in the current TurnIntents batch applied; continue same batch.
	advanceBatchIntentDone
)

// agentLoopState is mutable state for one Execute() run (one voice.transcript).
type agentLoopState struct {
	gathered agentcontext.Gathered
	// completed: intents already turned into directives in this Execute (merged with host-reported
	// successes when building [agentcontext.TurnContext].SucceededIntents).
	completed             []intents.Intent
	failedIntents         []agentcontext.FailedIntent
	contextRounds         int
	consecutiveContextReq int
	editCounter           int
	directives            []protocol.VoiceTranscriptDirective
	batchSourceIntents    []intents.Intent
	transcriptSummary     string
	// transcriptOutcome is protocol "transcriptOutcome": set to "irrelevant" when the agent returns TurnIrrelevant; empty otherwise.
	transcriptOutcome string
	maxRetries        int
}
