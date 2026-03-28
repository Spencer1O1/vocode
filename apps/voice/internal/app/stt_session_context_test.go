package app

import "testing"

func TestElevenLabsPreviousText_PrefixOnly(t *testing.T) {
	w := newUtteranceWindow(4, 1200)
	got := elevenLabsPreviousText(w)
	if got != elevenLabsSTTSessionContextPrefix {
		t.Fatalf("expected prefix only, got %q", got)
	}
}

func TestElevenLabsPreviousText_WithUtterances(t *testing.T) {
	w := newUtteranceWindow(4, 1200)
	w.AddUtterance("add a test")
	got := elevenLabsPreviousText(w)
	want := elevenLabsSTTSessionContextPrefix + " add a test"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
