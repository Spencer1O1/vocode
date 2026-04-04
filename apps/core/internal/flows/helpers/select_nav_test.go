package helpers

import "testing"

func TestParseNav_tabTwoScreenNotPickTwo(t *testing.T) {
	// "two" is part of "TabTwo" / "tab two screen", not list index 2.
	_, _, ok := ParseNav("go to the tab two screen component")
	if ok {
		t.Fatalf("expected no list-nav match for name-like phrase, got ok=true")
	}
}

func TestParseNav_goToTwoIsPick(t *testing.T) {
	k, ord, ok := ParseNav("go to two")
	if !ok || k != "pick" || ord != 2 {
		t.Fatalf("got ok=%v k=%q ord=%d", ok, k, ord)
	}
}

func TestParseNav_secondHit(t *testing.T) {
	k, ord, ok := ParseNav("go to the second hit")
	if !ok || k != "pick" || ord != 2 {
		t.Fatalf("got ok=%v k=%q ord=%d", ok, k, ord)
	}
}

func TestParseNav_next(t *testing.T) {
	k, _, ok := ParseNav("next")
	if !ok || k != "next" {
		t.Fatalf("got ok=%v k=%q", ok, k)
	}
}
