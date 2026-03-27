package symbols

import "testing"

func TestBuildParseSymbolID_RoundTrip(t *testing.T) {
	t.Parallel()

	in := SymbolRef{
		Name: "myFunction",
		Path: "/tmp/project/file.ts",
		Line: 42,
		Kind: "function",
	}
	id := BuildSymbolID(in)
	got, err := ParseSymbolID(id)
	if err != nil {
		t.Fatalf("expected parse success, got err: %v", err)
	}
	if got.Path != in.Path || got.Name != in.Name || got.Line != in.Line || got.Kind != in.Kind {
		t.Fatalf("round trip mismatch: got=%+v in=%+v", got, in)
	}
}

func TestParseSymbolID_Invalid(t *testing.T) {
	t.Parallel()
	if _, err := ParseSymbolID("bad-id"); err == nil {
		t.Fatal("expected parse error for malformed symbol id")
	}
}
