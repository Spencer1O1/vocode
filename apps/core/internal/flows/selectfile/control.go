package selectfileflow

import (
	"regexp"
	"strings"
)

var (
	selectionNavNextRe = regexp.MustCompile(`\b(next|forward)\b`)
	selectionNavBackRe = regexp.MustCompile(`\b(back|prev|previous)\b`)
	selectionNavGotoRe = regexp.MustCompile(`\b(goto|jump|show)\b`)
)

// Heuristic-based determination of control intent.
func parseControl(transcript string) (kind string, ok bool) {
	t := strings.TrimSpace(strings.ToLower(transcript))
	if selectionNavNextRe.MatchString(t) {
		return "next", true
	}
	if selectionNavBackRe.MatchString(t) {
		return "back", true
	}
	if selectionNavGotoRe.MatchString(t) {
		return "goto", true
	}
	return "", false
}

// HandleControl determines a control intent and handles it.
func HandleControl(transcript string) {
	intent, ok := parseControl(transcript)
	if !ok {
		return
	}

	switch intent {
	case "next":
		// Handle next
	case "back":
		// Handle back
	case "goto":
		// Handle goto by number (extract ordinal / cardinal from text)
	default:
		return
	}
}
