package fileselectflow

import "testing"

func TestFindFirstIdentifierOccurrence(t *testing.T) {
	src := "package foo\n\nfunc Bar() {}\n"
	line, char, ok := findFirstIdentifierOccurrence(src, "Bar")
	if !ok || line != 2 || char != 5 {
		t.Fatalf("Bar: got line=%d char=%d ok=%v", line, char, ok)
	}
	_, _, ok = findFirstIdentifierOccurrence(src, "Baz")
	if ok {
		t.Fatal("expected miss for Baz")
	}
}
