// Package stub provides a fixed-response [agent.ModelClient] for tests and dev wiring.
package stub

import (
	"context"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
)

// Client ignores most input and returns a deterministic turn sequence.
type Client struct{}

// New returns a [Client] that satisfies [agent.ModelClient].
func New() *Client {
	return &Client{}
}

// ScopedEdit implements [agent.ModelClient].
func (*Client) ScopedEdit(ctx context.Context, in agentcontext.ScopedEditContext) (agent.ScopedEditResult, error) {
	_ = ctx
	// Deterministic fixture for integration tests: if targetText contains the buggy comparator,
	// flip it. Otherwise, echo the original targetText (no-op style).
	text := in.TargetText
	if strings.Contains(text, "if (arr[j] < arr[j+1])") {
		text = strings.ReplaceAll(text, "if (arr[j] < arr[j+1])", "if (arr[j] > arr[j+1])")
	}
	return agent.ScopedEditResult{ReplacementText: text}, nil
}

func (*Client) ScopeIntent(ctx context.Context, in agentcontext.ScopeIntentContext) (agent.ScopeIntentResult, error) {
	_ = ctx
	// Simple deterministic stub: prefer current_function when we have a cursor symbol.
	if in.Editor.CursorSymbol != nil && strings.TrimSpace(in.Editor.CursorSymbol.Name) != "" {
		return agent.ScopeIntentResult{ScopeKind: agent.ScopeCurrentFunction}, nil
	}
	return agent.ScopeIntentResult{ScopeKind: agent.ScopeCurrentFile}, nil
}
