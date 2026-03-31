package hostcaps

import (
	"path/filepath"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/symbols"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// SymbolProvider is the daemon-side interface to host symbol capabilities.
// Today the VS Code extension provides LSP document symbols via VoiceTranscriptParams.ActiveFileSymbols.
// In the future this can be backed by a daemon-side LSP client without changing downstream consumers.
type SymbolProvider interface {
	ActiveFileSymbols(params protocol.VoiceTranscriptParams) []symbols.SymbolRef
}

// ParamsSymbolProvider adapts VoiceTranscriptParams.ActiveFileSymbols into SymbolRef values.
type ParamsSymbolProvider struct{}

func (ParamsSymbolProvider) ActiveFileSymbols(params protocol.VoiceTranscriptParams) []symbols.SymbolRef {
	active := strings.TrimSpace(params.ActiveFile)
	if active == "" || len(params.ActiveFileSymbols) == 0 {
		return nil
	}
	out := make([]symbols.SymbolRef, 0, len(params.ActiveFileSymbols))
	seen := map[string]bool{}
	for i := range params.ActiveFileSymbols {
		s := params.ActiveFileSymbols[i]
		ref := symbols.SymbolRef{
			Path: filepath.Clean(active),
			Line: int(s.SelectionRange.StartLine) + 1, // 1-based for v1 symbol ids
			Kind: strings.TrimSpace(s.Kind),
			Name: strings.TrimSpace(s.Name),
		}
		ref.ID = symbols.BuildSymbolID(ref)
		if ref.ID == "" || seen[ref.ID] {
			continue
		}
		seen[ref.ID] = true
		out = append(out, ref)
	}
	return out
}

