package tags

import (
	"math"
)

// SelectInnermostTag returns the tightest definition span that contains (line0, byteCol0) in tree-sitter
// coordinates. Tags without spans or non-definition tags are ignored.
func SelectInnermostTag(tags []Tag, line0, byteCol0 int) (Tag, bool) {
	var cands []Tag
	for i := range tags {
		t := tags[i]
		if !t.IsDefinition || !t.HasSpan() {
			continue
		}
		if t.Contains(line0, byteCol0) {
			cands = append(cands, t)
		}
	}
	if len(cands) == 0 {
		return Tag{}, false
	}
	return narrowestSpan(cands)
}

func narrowestSpan(cands []Tag) (Tag, bool) {
	if len(cands) == 0 {
		return Tag{}, false
	}
	best := cands[0]
	bestSize := spanSize(best)
	for i := 1; i < len(cands); i++ {
		sz := spanSize(cands[i])
		if sz < bestSize {
			bestSize = sz
			best = cands[i]
		}
	}
	return best, true
}

func spanSize(t Tag) int64 {
	if !t.HasSpan() {
		return math.MaxInt64
	}
	dLine := t.EndLine - t.StartLine
	dChar := t.EndCharacter - t.StartCharacter
	if dLine == 0 {
		return int64(dChar)
	}
	return int64(dLine)*1_000_000 + int64(dChar)
}
