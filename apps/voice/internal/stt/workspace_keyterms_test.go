package stt

import (
	"strings"
	"testing"
)

func TestMergeKeytermLists_dedupesCaseInsensitive(t *testing.T) {
	base := []string{"Foo", "bar"}
	extra := []string{"foo", "baz", " Bar "}
	got := mergeKeytermLists(base, extra)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3: %q", len(got), got)
	}
	if got[0] != "Foo" || got[1] != "bar" || got[2] != "baz" {
		t.Fatalf("unexpected order/content: %q", got)
	}
}

func TestNormalizeWorkspaceKeytermStrings_truncatesAndSkipsLongPhrases(t *testing.T) {
	long := strings.Repeat("x", 60)
	in := []string{"  ok  ", long, "one two three four five six", "dup", "Dup"}
	got := normalizeWorkspaceKeytermStrings(in)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3: %q", len(got), got)
	}
	if got[0] != "ok" {
		t.Fatalf("first=%q", got[0])
	}
	if len([]rune(got[1])) != maxKeytermRunes {
		t.Fatalf("truncated len runes=%d", len([]rune(got[1])))
	}
	if got[2] != "dup" {
		t.Fatalf("third=%q want dup", got[2])
	}
}

func TestNormalizeWorkspaceKeytermStrings_invalidJSONHandledByCaller(t *testing.T) {
	// parseEnv path uses jsonStringArray; empty on bad JSON
	if got := jsonStringArray("not json"); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
