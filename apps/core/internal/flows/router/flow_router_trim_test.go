package router

import "testing"

func TestTrimClassifierJSONResponse_plainJSON(t *testing.T) {
	in := `  {"route":"question","search_query":"","search_symbol_kind":""}  `
	want := `{"route":"question","search_query":"","search_symbol_kind":""}`
	if got := trimClassifierJSONResponse(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTrimClassifierJSONResponse_markdownFence(t *testing.T) {
	in := "```json\n{\"route\":\"create\",\"search_query\":\"\",\"search_symbol_kind\":\"\"}\n```"
	want := `{"route":"create","search_query":"","search_symbol_kind":""}`
	if got := trimClassifierJSONResponse(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTrimClassifierJSONResponse_leadingProse(t *testing.T) {
	in := `Here is the JSON:
{"route":"command","search_query":"","search_symbol_kind":""}`
	want := `{"route":"command","search_query":"","search_symbol_kind":""}`
	if got := trimClassifierJSONResponse(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
