package app

import (
	"os"
	"strings"
)

func sttEnabled() bool {
	// Default enabled to preserve existing behavior.
	v := strings.TrimSpace(os.Getenv("VOCODE_VOICE_STT_ENABLED"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	case "0", "false", "no", "n", "off", "disabled":
		return false
	default:
		// Fail open to avoid confusing "no transcripts" because of a typo.
		return true
	}
}

func sttMode() string {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("VOCODE_VOICE_STT_MODE")))
	switch v {
	case "", "batch":
		return "batch"
	case "stream", "streaming", "websocket", "ws":
		return "stream"
	default:
		return "batch"
	}
}
