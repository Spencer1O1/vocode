package searchapply

import (
	"reflect"
	"testing"
)

func TestContentSearchRgVariants(t *testing.T) {
	got := ContentSearchRgVariants("delta time")
	want := []string{"delta time", "deltaTime", "deltatime"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ContentSearchRgVariants(%q) = %#v, want %#v", "delta time", got, want)
	}
	if g := ContentSearchRgVariants("  foo  "); len(g) != 1 || g[0] != "foo" {
		t.Fatalf("single token: %#v", g)
	}
}
