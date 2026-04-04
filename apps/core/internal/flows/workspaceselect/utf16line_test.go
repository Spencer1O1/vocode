package workspaceselectflow

import "testing"

func TestLineUTF16Len_emoji(t *testing.T) {
	// a (1) + 😀 (2 UTF-16) + b (1) = 4
	s := "a😀b"
	if got := LineUTF16Len(s); got != 4 {
		t.Fatalf("LineUTF16Len: got %d want 4", got)
	}
}

func TestUTF16ColToByteOffset_emojiSlice(t *testing.T) {
	s := "a😀b"
	start := UTF16ColToByteOffset(s, 1)
	end := UTF16ColToByteOffset(s, 3)
	if got := s[start:end]; got != "😀" {
		t.Fatalf("slice: got %q want single emoji", got)
	}
}

func TestExtractRangeText_UTF16Columns(t *testing.T) {
	// Line 0: a😀b — select only the emoji (UTF-16 cols 1..3)
	body := "a😀b\n"
	got, ok := extractRangeText(body, 0, 1, 0, 3)
	if !ok {
		t.Fatal("extractRangeText: ok=false")
	}
	if got != "😀" {
		t.Fatalf("got %q want emoji", got)
	}
}
