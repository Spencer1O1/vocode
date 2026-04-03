package pipeline

import (
	"strings"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
	rootflow "vocoding.net/vocode/v2/apps/core/internal/flows/root"
	"vocoding.net/vocode/v2/apps/core/internal/flows/router"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/outcome"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/run"
	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func runMainPhase(
	e *run.Env,
	key string,
	params protocol.VoiceTranscriptParams,
	vs *session.VoiceSession,
	text string,
	pre preOpts,
) (protocol.VoiceTranscriptCompletion, bool, string) {
	if pre.has && pre.flow == flows.Root {
		rootD := rootDeps(e)
		fr := router.Result{Flow: flows.Root, Route: pre.route, SearchQuery: pre.searchQuery}
		res, fail := rootflow.DispatchRoute(rootD, params, vs, text, fr)
		if strings.TrimSpace(fail) != "" {
			persist(e, key, *vs)
			return protocol.VoiceTranscriptCompletion{Success: false}, true, fail
		}
		outcome.Apply(vs, params, res)
		persist(e, key, *vs)
		return res, true, ""
	}

	execRes, failure := rootflow.ExecuteMainPhase(rootDeps(e), params, vs, text)
	if strings.TrimSpace(failure) != "" {
		persist(e, key, *vs)
		return protocol.VoiceTranscriptCompletion{Success: false}, true, failure
	}
	outcome.Apply(vs, params, execRes)
	persist(e, key, *vs)
	return execRes, true, ""
}
