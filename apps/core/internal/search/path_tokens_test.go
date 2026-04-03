package search

import "testing"

func TestNormalizePathTokenForMatch(t *testing.T) {
	if a, b := NormalizePathTokenForMatch("my_utils"), NormalizePathTokenForMatch("My Utils"); a != b {
		t.Fatalf("my_utils %q vs My Utils %q", a, b)
	}
	if NormalizePathTokenForMatch("pkg-name") != NormalizePathTokenForMatch("PkgName") {
		t.Fatal("pkg-name vs PkgName")
	}
}
