package run

import (
	"vocoding.net/vocode/v2/apps/core/internal/flows/router"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/searchapply"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Env holds wiring for transcript execution (sessions, host, router, search apply).
type Env struct {
	Sessions   *session.VoiceSessionStore
	Ephemeral  *session.VoiceSession
	HostApply  HostApplyClient
	FlowRouter *router.FlowRouter
	Search     *searchapply.TranscriptSearch
}

type HostApplyClient interface {
	ApplyDirectives(protocol.HostApplyParams) (protocol.HostApplyResult, error)
}
