package app

import "testing"

func TestUtteranceWindow_Basic(t *testing.T) {
	w := newUtteranceWindow(3, 1000)
	w.AddUtterance("hello")
	w.AddUtterance("world")
	if got := w.PreviousText(); got != "hello world" {
		t.Fatalf("unexpected previous text: %q", got)
	}
}

func TestUtteranceWindow_RespectsUtteranceCount(t *testing.T) {
	w := newUtteranceWindow(2, 1000)
	w.AddUtterance("one")
	w.AddUtterance("two")
	w.AddUtterance("three")
	if got := w.PreviousText(); got != "two three" {
		t.Fatalf("expected last two utterances, got %q", got)
	}
}

func TestUtteranceWindow_RespectsCharLimit(t *testing.T) {
	w := newUtteranceWindow(5, 12)
	w.AddUtterance("first")
	w.AddUtterance("second")
	w.AddUtterance("third")
	if got := w.PreviousText(); got != "second third" {
		t.Fatalf("expected oldest utterance dropped by char cap, got %q", got)
	}
}
