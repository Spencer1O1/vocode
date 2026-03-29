// Package voicesession loads and saves per-context transcript state: [agentcontext.VoiceSession]
// in [agentcontext.VoiceSessionStore], plus process-local pending directive batch when
// params.contextSessionId is empty.
package voicesession

import (
	"fmt"
	"strings"
	"time"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Load returns session state for this RPC. When contextKey is empty, only the ephemeral
// pending directive batch is restored (gathered always starts empty for that mode).
func Load(store *agentcontext.VoiceSessionStore, contextKey string, idleReset time.Duration, ephemeralPending *agentcontext.DirectiveApplyBatch) agentcontext.VoiceSession {
	key := strings.TrimSpace(contextKey)
	if key == "" {
		if ephemeralPending == nil {
			return agentcontext.VoiceSession{}
		}
		p := *ephemeralPending
		p.SourceIntents = append([]intents.Intent(nil), ephemeralPending.SourceIntents...)
		return agentcontext.VoiceSession{PendingDirectiveApply: &p}
	}
	return store.Get(key, idleReset)
}

// SaveKeyed persists vs when contextKey is non-empty.
func SaveKeyed(store *agentcontext.VoiceSessionStore, contextKey string, vs agentcontext.VoiceSession) {
	key := strings.TrimSpace(contextKey)
	if key == "" || store == nil {
		return
	}
	store.Put(key, vs)
}

// StoreEphemeralPending copies vs.PendingDirectiveApply into *dst (nil when none).
func StoreEphemeralPending(dst **agentcontext.DirectiveApplyBatch, vs agentcontext.VoiceSession) {
	if vs.PendingDirectiveApply == nil {
		*dst = nil
		return
	}
	p := *vs.PendingDirectiveApply
	p.SourceIntents = append([]intents.Intent(nil), vs.PendingDirectiveApply.SourceIntents...)
	*dst = &p
}

// ConsumeIncomingApplyReport strips apply-report fields from params and updates vs.
func ConsumeIncomingApplyReport(params *protocol.VoiceTranscriptParams, vs *agentcontext.VoiceSession) ([]intents.Intent, []agentcontext.FailedIntent, error) {
	items := params.LastBatchApply
	reportID := strings.TrimSpace(params.ReportApplyBatchId)
	params.LastBatchApply = nil
	params.ReportApplyBatchId = ""

	if len(items) == 0 {
		vs.PendingDirectiveApply = nil
		return nil, nil, nil
	}
	if vs.PendingDirectiveApply == nil {
		return nil, nil, fmt.Errorf("lastBatchApply without pending directive apply batch")
	}
	extSucc, extFail, err := vs.PendingDirectiveApply.ConsumeHostApplyReport(reportID, items)
	if err != nil {
		return nil, nil, err
	}
	vs.PendingDirectiveApply = nil
	return extSucc, extFail, nil
}
