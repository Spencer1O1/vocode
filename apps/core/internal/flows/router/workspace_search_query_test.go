package router

import (
	"testing"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

func TestNormalizeWorkspaceSelectSearchQuery_stripsTrailingComponent(t *testing.T) {
	if g := NormalizeWorkspaceSelectSearchQuery("test component"); g != "test" {
		t.Fatalf("got %q want test", g)
	}
}

func TestNormalizeWorkspaceSelectSearchQuery_stripsChainedSuffixes(t *testing.T) {
	if g := NormalizeWorkspaceSelectSearchQuery("test component."); g != "test" {
		t.Fatalf("got %q want test", g)
	}
}

func TestNormalizeWorkspaceSelectSearchQuery_singleTokenUnchanged(t *testing.T) {
	if g := NormalizeWorkspaceSelectSearchQuery("symbol"); g != "symbol" {
		t.Fatalf("got %q want symbol", g)
	}
}

func TestNormalizeWorkspaceSelectSearchQuery_multiWordNameKept(t *testing.T) {
	if g := NormalizeWorkspaceSelectSearchQuery("parse header"); g != "parse header" {
		t.Fatalf("got %q", g)
	}
}

func TestDisambiguateClassifierResult_normalizesWorkspaceQuery(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find the test component"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "test component"}
	got := DisambiguateClassifierResult(in, res)
	if got.SearchQuery != "test" {
		t.Fatalf("got %q want test", got.SearchQuery)
	}
}
