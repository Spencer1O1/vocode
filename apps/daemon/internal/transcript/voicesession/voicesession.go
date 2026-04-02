// Package voicesession loads and saves per-context transcript state: [agentcontext.VoiceSession]
// in [agentcontext.VoiceSessionStore], plus process-local pending directive batch when
// params.contextSessionId is empty.
package voicesession

import (
	"fmt"
	"strings"
	"time"

	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Load returns session state for this RPC. When contextKey is empty, restores the full
// ephemeral [agentcontext.VoiceSession] (gathered, apply history, pending batch) from *ephemeral.
func Load(store *agentcontext.VoiceSessionStore, contextKey string, idleReset time.Duration, ephemeral *agentcontext.VoiceSession) agentcontext.VoiceSession {
	key := strings.TrimSpace(contextKey)
	if key == "" {
		if ephemeral == nil {
			return agentcontext.VoiceSession{}
		}
		// Deep-copy slices and batch pointer so RPC-local mutations cannot alias the
		// stored ephemeral buffer (see [agentcontext.CloneVoiceSession]).
		return agentcontext.CloneVoiceSession(*ephemeral)
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

// StoreEphemeralVoiceSession copies vs into *dst for the next RPC when contextSessionId is empty.
func StoreEphemeralVoiceSession(dst *agentcontext.VoiceSession, vs agentcontext.VoiceSession) {
	*dst = vs
}

// ConsumeHostApplyReport consumes the host's apply outcomes for the currently
// pending directive batch, updates intent apply history, and clears the pending
// batch on success.
func ConsumeHostApplyReport(
	reportID string,
	items []protocol.VoiceTranscriptDirectiveApplyItem,
	vs *agentcontext.VoiceSession,
) error {
	if len(items) == 0 {
		vs.PendingDirectiveApply = nil
		return nil
	}
	if vs.PendingDirectiveApply == nil {
		return fmt.Errorf("host apply report without pending directive apply batch")
	}
	batch := vs.PendingDirectiveApply
	if err := batch.ConsumeHostApplyReport(reportID, items); err != nil {
		return err
	}
	for i := range items {
		if strings.TrimSpace(items[i].Status) == agentcontext.ApplyItemStatusFailed {
			msg := strings.TrimSpace(items[i].Message)
			if msg == "" {
				msg = "host apply failed"
			}
			return fmt.Errorf("%s", msg)
		}
	}
	vs.PendingDirectiveApply = nil
	return nil
}
