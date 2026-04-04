package helpers

import (
	"regexp"
	"strings"
)

var (
	navNextRe = regexp.MustCompile(`\b(next|forward)\b`)
	navBackRe = regexp.MustCompile(`\b(back|prev|previous)\b`)
	navGotoRe = regexp.MustCompile(`\b(goto|jump|show)\b`)
	navIntRe  = regexp.MustCompile(`\b\d+\b`)
)

// ParseNav parses list-navigation intent from a voice transcript.
// kind is one of: "next", "back", "pick", "exit". ord is 1-based for "pick".
func ParseNav(transcript string) (kind string, ord int, ok bool) {
	t := strings.TrimSpace(transcript)
	if t == "" {
		return "", 0, false
	}
	if IsExitPhrase(t) {
		return "exit", 0, true
	}
	lower := strings.ToLower(t)
	if navNextRe.MatchString(lower) {
		return "next", 0, true
	}
	if navBackRe.MatchString(lower) {
		return "back", 0, true
	}
	if n := parsePickOrdinal(t); n > 0 {
		return "pick", n, true
	}
	if navGotoRe.MatchString(lower) {
		if n := parseAnyIntToken(t); n > 0 {
			return "pick", n, true
		}
	}
	return "", 0, false
}

// pickOrdinalNeighborHint is true for words that suggest the previous/next token is a list index, not part of a symbol name (e.g. "tab two screen" spoken for TabTwo).
func pickOrdinalNeighborHint(tok string) bool {
	switch strings.Trim(tok, ".,;:!?") {
	case "the", "a", "an", "hit", "hits", "result", "results", "match", "matches",
		"item", "items", "entry", "entries", "number", "pick", "row", "rows", "choice", "choices":
		return true
	default:
		return false
	}
}

// ordinalPickContext reports whether a spoken ordinal at fields[idx] is meant as list position, not a syllable inside a name ("… tab two screen …").
func ordinalPickContext(fields []string, idx int) bool {
	n := len(fields)
	if n <= 3 {
		return true
	}
	if idx == n-1 {
		return true
	}
	prev := ""
	if idx > 0 {
		prev = strings.ToLower(strings.Trim(fields[idx-1], ".,;:!?"))
	}
	next := ""
	if idx+1 < n {
		next = strings.ToLower(strings.Trim(fields[idx+1], ".,;:!?"))
	}
	return pickOrdinalNeighborHint(prev) || pickOrdinalNeighborHint(next)
}

func parsePickOrdinal(s string) int {
	if n := parseAnyIntToken(s); n > 0 {
		if navIntRe.MatchString(strings.TrimSpace(s)) && len(strings.Fields(strings.TrimSpace(s))) == 1 {
			return n
		}
	}
	t := strings.TrimSpace(strings.ToLower(s))
	fields := strings.Fields(t)
	for i, raw := range fields {
		w := strings.Trim(raw, ".,;:!?")
		var ord int
		switch w {
		case "one", "1st", "first":
			ord = 1
		case "two", "2nd", "second":
			ord = 2
		case "three", "3rd", "third":
			ord = 3
		case "four", "4th", "fourth":
			ord = 4
		case "five", "5th", "fifth":
			ord = 5
		case "six", "6th", "sixth":
			ord = 6
		case "seven", "7th", "seventh":
			ord = 7
		case "eight", "8th", "eighth":
			ord = 8
		case "nine", "9th", "ninth":
			ord = 9
		case "ten", "10th", "tenth":
			ord = 10
		default:
			continue
		}
		if ordinalPickContext(fields, i) {
			return ord
		}
	}
	return 0
}

func parseAnyIntToken(s string) int {
	n := 0
	inDigits := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			inDigits = true
			n = n*10 + int(c-'0')
			continue
		}
		if inDigits {
			return n
		}
	}
	if inDigits {
		return n
	}
	return 0
}
