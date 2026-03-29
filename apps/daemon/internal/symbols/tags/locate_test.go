package tags

import "testing"

func TestSelectInnermostTag_nestedSpans(t *testing.T) {
	t.Parallel()
	tags := []Tag{
		{
			Name: "outer", Kind: "function", IsDefinition: true,
			StartLine: 0, StartCharacter: 0, EndLine: 20, EndCharacter: 1,
		},
		{
			Name: "inner", Kind: "function", IsDefinition: true,
			StartLine: 2, StartCharacter: 0, EndLine: 8, EndCharacter: 1,
		},
	}
	got, ok := SelectInnermostTag(tags, 5, 0)
	if !ok || got.Name != "inner" {
		t.Fatalf("inside inner span: got (%v, %v), want inner", got, ok)
	}
}

func TestSelectInnermostTag_byteColumn(t *testing.T) {
	t.Parallel()
	tags := []Tag{
		{
			Name: "wide", Kind: "function", IsDefinition: true,
			StartLine: 0, StartCharacter: 0, EndLine: 0, EndCharacter: 20,
		},
	}
	got, ok := SelectInnermostTag(tags, 0, 5)
	if !ok || got.Name != "wide" {
		t.Fatalf("inside columns: got (%v, %v)", got, ok)
	}
	_, ok = SelectInnermostTag(tags, 0, 25)
	if ok {
		t.Fatal("past end column should not match")
	}
}

func TestSelectInnermostTag_noSpanNoMatch(t *testing.T) {
	t.Parallel()
	tags := []Tag{
		{Name: "nospan", Kind: "function", IsDefinition: true, Line: 5},
	}
	_, ok := SelectInnermostTag(tags, 4, 0)
	if ok {
		t.Fatal("tag without geometry must not match")
	}
}
