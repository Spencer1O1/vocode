package router

import (
	"testing"

	"vocoding.net/vocode/v2/apps/core/internal/flows"
)

func TestDisambiguateClassifierResult_noPathWords_workspaceSelect(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find the test component"}
	res := Result{Flow: flows.Root, Route: "file_select", SearchQuery: "test"}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "workspace_select" || got.SearchQuery != "test" {
		t.Fatalf("got %+v", got)
	}
}

func TestDisambiguateClassifierResult_pathWord_fileSelect(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find the test file"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "test"}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "file_select" || got.SearchQuery != "test" {
		t.Fatalf("got %+v", got)
	}
}

func TestDisambiguateClassifierResult_openVerb_fileSelect(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "open readme"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "readme"}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "file_select" {
		t.Fatalf("got route %q want file_select", got.Route)
	}
}

func TestDisambiguateClassifierResult_folder_fileSelect(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find the utils folder"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "utils"}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "file_select" {
		t.Fatalf("got route %q want file_select", got.Route)
	}
}

func TestDisambiguateClassifierResult_pathWord_clearsSymbolKind(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find config file"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "config", SearchSymbolKind: "variable"}
	got := DisambiguateClassifierResult(in, res)
	if got.SearchSymbolKind != "" {
		t.Fatalf("want empty symbol kind, got %q", got.SearchSymbolKind)
	}
}

func TestDisambiguateClassifierResult_workspaceKeepsSymbolKind(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "go to main"}
	res := Result{Flow: flows.Root, Route: "workspace_select", SearchQuery: "main", SearchSymbolKind: "function"}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "workspace_select" || got.SearchSymbolKind != "function" {
		t.Fatalf("got %+v", got)
	}
}

func TestDisambiguateClassifierResult_otherRouteUnchanged(t *testing.T) {
	in := Context{Flow: flows.Root, Instruction: "find the test file"}
	res := Result{Flow: flows.Root, Route: "command", SearchQuery: ""}
	got := DisambiguateClassifierResult(in, res)
	if got.Route != "command" {
		t.Fatalf("got route %q", got.Route)
	}
}
