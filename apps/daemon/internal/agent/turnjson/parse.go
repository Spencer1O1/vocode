// Package turnjson parses model JSON into [agent.TurnResult]. Daemon-owned wire shape (not protocol).
package turnjson

import (
	"encoding/json"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
)

// wireTurn is the model output envelope (single JSON object per completion).
type wireTurn struct {
	Kind          string                          `json:"kind"`
	Reason        string                          `json:"reason,omitempty"`
	Summary       string                          `json:"summary,omitempty"`
	GatherContext *agentcontext.GatherContextSpec `json:"requestContext,omitempty"`
	Intents       []json.RawMessage               `json:"intents,omitempty"`
}

// ParseTurn unmarshals one JSON object into [agent.TurnResult] and validates it.
func ParseTurn(data []byte) (agent.TurnResult, error) {
	data = trimJSONFences(data)
	var w wireTurn
	if err := json.Unmarshal(data, &w); err != nil {
		return agent.TurnResult{}, fmt.Errorf("turn json: %w", err)
	}
	k := strings.TrimSpace(strings.ToLower(w.Kind))
	switch k {
	case string(agent.TurnIrrelevant):
		t := agent.TurnResult{
			Kind:             agent.TurnIrrelevant,
			IrrelevantReason: strings.TrimSpace(w.Reason),
		}
		if err := t.Validate(); err != nil {
			return agent.TurnResult{}, err
		}
		return t, nil
	case string(agent.TurnFinish):
		t := agent.TurnResult{
			Kind:          agent.TurnFinish,
			FinishSummary: strings.TrimSpace(w.Summary),
		}
		if err := t.Validate(); err != nil {
			return agent.TurnResult{}, err
		}
		return t, nil
	case string(agent.TurnGatherContext):
		if w.GatherContext == nil {
			return agent.TurnResult{}, fmt.Errorf("turn json: request_context requires requestContext")
		}
		t := agent.TurnResult{
			Kind:          agent.TurnGatherContext,
			GatherContext: w.GatherContext,
		}
		if err := t.Validate(); err != nil {
			return agent.TurnResult{}, err
		}
		return t, nil
	case string(agent.TurnIntents):
		if len(w.Intents) == 0 {
			return agent.TurnResult{}, fmt.Errorf("turn json: intents requires non-empty intents array")
		}
		out := make([]intents.Intent, 0, len(w.Intents))
		for i, raw := range w.Intents {
			var it intents.Intent
			if err := json.Unmarshal(raw, &it); err != nil {
				return agent.TurnResult{}, fmt.Errorf("turn json: intents[%d]: %w", i, err)
			}
			out = append(out, it)
		}
		t := agent.TurnResult{Kind: agent.TurnIntents, Intents: out}
		if err := t.Validate(); err != nil {
			return agent.TurnResult{}, err
		}
		return t, nil
	default:
		return agent.TurnResult{}, fmt.Errorf("turn json: unknown kind %q", w.Kind)
	}
}

func trimJSONFences(b []byte) []byte {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			if idx := strings.IndexByte(s, '\n'); idx >= 0 {
				s = strings.TrimSpace(s[idx+1:])
			}
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return []byte(s)
}
