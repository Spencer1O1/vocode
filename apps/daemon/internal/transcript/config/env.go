// Package config reads environment variables for the voice transcript daemon path.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Int parses an int from the environment, or returns def when missing or invalid.
func Int(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

const defaultSessionIdleReset = 30 * time.Minute

// SessionIdleReset returns idle TTL for dropping stored voice session rows (see agentcontext.VoiceSessionStore).
// Unset or invalid → 30m default. Explicit "0" → disabled (no idle eviction).
func SessionIdleReset() time.Duration {
	v := strings.TrimSpace(os.Getenv("VOCODE_DAEMON_SESSION_IDLE_RESET_MS"))
	if v == "" {
		return defaultSessionIdleReset
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultSessionIdleReset
	}
	if i == 0 {
		return 0
	}
	if i < 0 {
		return defaultSessionIdleReset
	}
	return time.Duration(i) * time.Millisecond
}
