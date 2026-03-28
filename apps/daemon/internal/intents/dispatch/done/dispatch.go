package done

import (
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
)

// Dispatch maps a validated done intent payload to the trimmed summary string for the host
// (copied onto VoiceTranscriptResult.summary by the transcript executor). Nil payload is valid.
func Dispatch(d *intents.DoneIntent) (summary string, err error) {
	if d == nil {
		return "", nil
	}
	return strings.TrimSpace(d.Summary), nil
}
