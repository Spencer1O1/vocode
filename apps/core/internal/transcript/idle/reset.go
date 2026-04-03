package idle

import (
	"time"

	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

func SessionResetDuration(params protocol.VoiceTranscriptParams) time.Duration {
	if params.DaemonConfig == nil || params.DaemonConfig.SessionIdleResetMs == nil {
		return 0
	}
	ms := *params.DaemonConfig.SessionIdleResetMs
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}
