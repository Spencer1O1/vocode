package stt

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

// ElevenLabs keyterm limits (batch STT docs; align realtime requests conservatively).
const (
	maxWorkspaceKeytermCount = 80
	maxKeytermRunes          = 50
	maxKeytermWords          = 5
)

func parseEnvWorkspaceKeyterms() []string {
	raw := strings.TrimSpace(os.Getenv("VOCODE_STT_KEYTERMS_JSON"))
	if raw == "" {
		return nil
	}
	return normalizeWorkspaceKeytermStrings(jsonStringArray(raw))
}

func jsonStringArray(raw string) []string {
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil
	}
	return arr
}

func normalizeWorkspaceKeytermStrings(arr []string) []string {
	if len(arr) == 0 {
		return nil
	}
	out := make([]string, 0, len(arr))
	seen := make(map[string]struct{}, len(arr))
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if utf8.RuneCountInString(s) > maxKeytermRunes {
			s = truncateRunes(s, maxKeytermRunes)
		}
		if len(strings.Fields(s)) > maxKeytermWords {
			continue
		}
		k := strings.ToLower(s)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
		if len(out) >= maxWorkspaceKeytermCount {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func truncateRunes(s string, max int) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func mergeKeytermLists(base, extra []string) []string {
	seen := make(map[string]struct{}, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))
	for _, s := range base {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		k := strings.ToLower(s)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
	}
	for _, s := range extra {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		k := strings.ToLower(s)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
	}
	return out
}

func mergedRealtimeSTTKeytermsUncached() []string {
	return mergeKeytermLists(RealtimeSTTKeyterms, parseEnvWorkspaceKeyterms())
}

var mergedRealtimeKeytermsOnce sync.Once
var mergedRealtimeKeyterms []string

// AllRealtimeSTTKeyterms returns built-in keyterms plus any from VOCODE_STT_KEYTERMS_JSON
// (JSON string array), merged and de-duplicated case-insensitively. Workspace extras are appended
// after built-ins. Parsed once per process.
func AllRealtimeSTTKeyterms() []string {
	mergedRealtimeKeytermsOnce.Do(func() {
		mergedRealtimeKeyterms = mergedRealtimeSTTKeytermsUncached()
	})
	return mergedRealtimeKeyterms
}
