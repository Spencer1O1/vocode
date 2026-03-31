package executor

import (
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/gather"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch"
	"vocoding.net/vocode/v2/apps/daemon/internal/symbols"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Executor runs one voice.transcript through the agent: an iterative loop of [agent.Agent.NextTurn],
// optional gather-context rounds between batches, batched executable intents per turn, retries,
// [gather.FulfillSpec] for turn-level context enrichment, and [dispatch.Handler.Handle] for host directives.
type Executor struct {
	agent                    *agent.Agent
	intentHandler            *dispatch.Handler
	gather                   *gather.Provider
	symbols                  symbols.Resolver
	maxAgentTurns            int
	maxIntentRetries         int
	maxContextRounds         int
	maxContextBytes          int
	maxConsecutiveContextReq int
	maxIntentsPerBatch       int
}

type ExecutionCaps struct {
	MaxAgentTurns            int
	MaxIntentRetries         int
	MaxContextRounds         int
	MaxContextBytes          int
	MaxConsecutiveContextReq int
	MaxIntentsPerBatch       int
}

func (e *Executor) effectiveCaps(params protocol.VoiceTranscriptParams) ExecutionCaps {
	caps := ExecutionCaps{
		MaxAgentTurns:            e.maxAgentTurns,
		MaxIntentRetries:         e.maxIntentRetries,
		MaxContextRounds:         e.maxContextRounds,
		MaxContextBytes:          e.maxContextBytes,
		MaxConsecutiveContextReq: e.maxConsecutiveContextReq,
		MaxIntentsPerBatch:       e.maxIntentsPerBatch,
	}

	dc := params.DaemonConfig
	if dc == nil {
		// Back-compat: keep env-derived defaults.
		return caps
	}

	if dc.MaxPlannerTurns != nil {
		caps.MaxAgentTurns = int(*dc.MaxPlannerTurns)
	}
	if dc.MaxIntentDispatchRetries != nil {
		caps.MaxIntentRetries = int(*dc.MaxIntentDispatchRetries)
	}
	if dc.MaxContextRounds != nil {
		caps.MaxContextRounds = int(*dc.MaxContextRounds)
	}
	if dc.MaxContextBytes != nil {
		caps.MaxContextBytes = int(*dc.MaxContextBytes)
	}
	if dc.MaxConsecutiveContextRequests != nil {
		caps.MaxConsecutiveContextReq = int(*dc.MaxConsecutiveContextRequests)
	}
	if dc.MaxIntentsPerBatch != nil {
		caps.MaxIntentsPerBatch = int(*dc.MaxIntentsPerBatch)
	}

	// Normalize to preserve existing executor semantics.
	if caps.MaxAgentTurns <= 0 {
		caps.MaxAgentTurns = 8
	}
	if caps.MaxIntentRetries < 0 {
		caps.MaxIntentRetries = 0
	}
	return caps
}

// Options configures caps and optional symbol resolution for [Executor].
type Options struct {
	MaxAgentTurns            int
	MaxIntentRetries         int
	MaxContextRounds         int
	MaxContextBytes          int
	MaxConsecutiveContextReq int
	// MaxIntentsPerBatch caps turn "intents" batch length; 0 or negative means no cap.
	MaxIntentsPerBatch int
	Symbols            symbols.Resolver
}

// New constructs an [Executor].
// MaxIntentsPerBatch: 0 means no cap; unset env defaults to 16 in [transcript.NewService].
func New(a *agent.Agent, h *dispatch.Handler, gatherProv *gather.Provider, opts Options) *Executor {
	if opts.Symbols == nil {
		opts.Symbols = symbols.NewTreeSitterResolver()
	}
	return &Executor{
		agent:                    a,
		intentHandler:            h,
		gather:                   gatherProv,
		symbols:                  opts.Symbols,
		maxAgentTurns:            opts.MaxAgentTurns,
		maxIntentRetries:         opts.MaxIntentRetries,
		maxContextRounds:         opts.MaxContextRounds,
		maxContextBytes:          opts.MaxContextBytes,
		maxConsecutiveContextReq: opts.MaxConsecutiveContextReq,
		maxIntentsPerBatch:       opts.MaxIntentsPerBatch,
	}
}

// Execute runs the agent loop until done, caps, or failure.
func (e *Executor) Execute(
	params protocol.VoiceTranscriptParams,
	gatheredIn agentcontext.Gathered,
	intentApplyHistory []agentcontext.IntentApplyRecord,
	extSucceeded []intents.Intent,
	extFailed []agentcontext.FailedIntent,
	extSkipped []intents.Intent,
) (protocol.VoiceTranscriptCompletion, []protocol.VoiceTranscriptDirective, agentcontext.Gathered, *agentcontext.DirectiveApplyBatch, bool, string) {
	text := strings.TrimSpace(params.Text)
	if text == "" {
		return protocol.VoiceTranscriptCompletion{}, nil, gatheredIn, nil, false, "empty transcript text"
	}
	caps := e.effectiveCaps(params)
	maxLoopIters := caps.MaxAgentTurns
	maxRetries := caps.MaxIntentRetries
	hostCursor := resolveHostCursorSymbol(e.symbols, params)
	st := &agentLoopState{
		gathered:   agentcontext.SeedGatheredActiveFile(gatheredIn, params.ActiveFile),
		completed:  make([]intents.Intent, 0, maxLoopIters),
		directives: make([]protocol.VoiceTranscriptDirective, 0, maxLoopIters*4),
		maxRetries: maxRetries,
	}
	// Pre-populate active-file symbols on turn 0 so the planner can use symbol_id targets
	// without immediately spending a context round on request_symbols.
	if e.symbols != nil && len(st.gathered.Symbols) == 0 && strings.TrimSpace(params.ActiveFile) != "" {
		syms, err := e.symbols.ResolveFileSymbols(strings.TrimSpace(params.WorkspaceRoot), strings.TrimSpace(params.ActiveFile))
		if err != nil {
			st.gathered = appendGatheredNote(st.gathered, fmt.Sprintf("symbol resolver unavailable for active file: %v", err))
		} else if len(syms) > 0 {
			// Merge by id.
			seen := map[string]bool{}
			for _, s := range st.gathered.Symbols {
				seen[s.ID] = true
			}
			for _, s := range syms {
				if s.ID == "" || seen[s.ID] {
					continue
				}
				seen[s.ID] = true
				st.gathered.Symbols = append(st.gathered.Symbols, s)
				if len(st.gathered.Symbols) >= 50 {
					break
				}
			}
		}
	}

	brokeOK := false
	for range maxLoopIters {
		adv, failRes, abort, reason := e.runOneAgentLoopIteration(
			params,
			text,
			hostCursor,
			intentApplyHistory,
			extSucceeded,
			extFailed,
			extSkipped,
			st,
			caps,
		)
		if abort {
			if strings.TrimSpace(reason) == "" {
				reason = "executor aborted"
			}
			return failRes, nil, st.gathered, nil, true, reason
		}
		if adv == advanceBreakLoop {
			brokeOK = true
			break
		}
	}

	if !brokeOK {
		return protocol.VoiceTranscriptCompletion{Success: false}, nil, st.gathered, nil, true, fmt.Sprintf("planner did not finish within maxPlannerTurns=%d", maxLoopIters)
	}
	res, dirs, g2, pending, ok := finalizeExecute(st)
	if !ok {
		return protocol.VoiceTranscriptCompletion{Success: false}, nil, g2, nil, true, "failed to finalize transcript result"
	}
	return res, dirs, g2, pending, ok, ""
}
