package agentcontext

import "vocoding.net/vocode/v2/apps/daemon/internal/symbols"

// FileExcerpt is a path plus text slice (typically from request_file_excerpt).
type FileExcerpt struct {
	Path    string
	Content string
}

// Gathered holds file excerpts, symbols, and notes assembled for TurnContext.
// The executor seeds the active file each RPC, loads prior state from [VoiceSessionStore] when
// contextSessionId is set, and grows entries across request_context turns; duplicate paths refresh in place.
type Gathered struct {
	Symbols  []symbols.SymbolRef
	Excerpts []FileExcerpt
	Notes    []string
}
