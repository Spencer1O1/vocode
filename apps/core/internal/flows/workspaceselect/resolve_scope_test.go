package workspaceselectflow

import (
	"testing"

	"vocoding.net/vocode/v2/apps/core/internal/transcript/session"
)

func TestRangeForSearchHit_singleLine(t *testing.T) {
	body := "hello world\n"
	sl, sc, el, ec, ok := RangeForSearchHit(body, session.SearchHit{Line: 0, Character: 6, Len: 5})
	if !ok || sl != 0 || sc != 6 || el != 0 || ec != 11 {
		t.Fatalf("got (%d,%d)-(%d,%d) ok=%v want 0,6-0,11", sl, sc, el, ec, ok)
	}
}

func TestRangeForSearchHit_crossNewline(t *testing.T) {
	body := "ab\nx\n"
	sl, sc, el, ec, ok := RangeForSearchHit(body, session.SearchHit{Line: 0, Character: 0, Len: 4})
	if !ok || sl != 0 || sc != 0 || el != 1 || ec != 1 {
		t.Fatalf("got (%d,%d)-(%d,%d) ok=%v want 0,0-1,1 for ab\\nx", sl, sc, el, ec, ok)
	}
}
