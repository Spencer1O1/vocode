package tags

import "testing"

func TestUtf16PrefixByteLength_ascii(t *testing.T) {
	t.Parallel()
	if got := utf16PrefixByteLength("hello", 3); got != 3 {
		t.Fatalf("got %d want 3", got)
	}
}

func TestUtf16PrefixByteLength_surrogatePair(t *testing.T) {
	t.Parallel()
	s := "a😀b"
	if got := utf16PrefixByteLength(s, 1); got != 1 {
		t.Fatalf("after one ASCII: got %d", got)
	}
	if got := utf16PrefixByteLength(s, 3); got != 1+4 {
		t.Fatalf("after emoji: got %d want 5", got)
	}
}
