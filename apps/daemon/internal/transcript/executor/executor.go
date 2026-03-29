package executor

import (
	"context"
	"fmt"
	"strings"

	"vocoding.net/vocode/v2/apps/daemon/internal/agent"
	"vocoding.net/vocode/v2/apps/daemon/internal/agentcontext"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch"
	"vocoding.net/vocode/v2/apps/daemon/internal/intents/dispatch/edit"
	"vocoding.net/vocode/v2/apps/daemon/internal/symbols"
	protocol "vocoding.net/vocode/v2/packages/protocol/go"
)

// Executor runs one voice.transcript through the agent: iterative NextIntent turns, optional
// request_context rounds, retries, and dispatch via [dispatch.Handler.Handle] (control vs executable).
type Executor struct {
	agent                    *agent.Agent
	intentHandler            *dispatch.Handler
	symbols                  symbols.Resolver
	maxAgentTurns            int
	maxIntentRetries         int
	maxContextRounds         int
	maxContextBytes          int
	maxConsecutiveContextReq int
}

// Options configures caps and optional symbol resolution for [Executor].
type Options struct {
	MaxAgentTurns            int
	MaxIntentRetries         int
	MaxContextRounds         int
	MaxContextBytes          int
	MaxConsecutiveContextReq int
	Symbols                  symbols.Resolver
}

// New constructs an [Executor].
func New(a *agent.Agent, h *dispatch.Handler, opts Options) *Executor {
	return &Executor{
		agent:                    a,
		intentHandler:            h,
		symbols:                  opts.Symbols,
		maxAgentTurns:            opts.MaxAgentTurns,
		maxIntentRetries:         opts.MaxIntentRetries,
		maxContextRounds:         opts.MaxContextRounds,
		maxContextBytes:          opts.MaxContextBytes,
		maxConsecutiveContextReq: opts.MaxConsecutiveContextReq,
	}
}

// Execute runs the planner loop until done, caps, or failure.
func (e *Executor) Execute(params protocol.VoiceTranscriptParams, gatheredIn agentcontext.Gathered, extSucceeded []intents.Intent, extFailed []agentcontext.FailedIntent) (protocol.VoiceTranscriptResult, agentcontext.Gathered, *agentcontext.DirectiveApplyBatch, bool) {
	text := strings.TrimSpace(params.Text)
	if text == "" {
		return protocol.VoiceTranscriptResult{}, gatheredIn, nil, false
	}
	maxTurns := e.maxAgentTurns
	if maxTurns <= 0 {
		maxTurns = 8
	}
	maxRetries := e.maxIntentRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	gathered := agentcontext.SeedGatheredActiveFile(gatheredIn, params.ActiveFile)
	hostCursor := resolveHostCursorSymbol(e.symbols, params)
	completed := make([]intents.Intent, 0, maxTurns)
	var failedIntents []agentcontext.FailedIntent
	contextRounds := 0
	consecutiveContextReq := 0
	editCounter := 0
	directives := make([]protocol.VoiceTranscriptDirective, 0, maxTurns)
	var batchSourceIntents []intents.Intent
	trace := make([]string, 0, maxTurns*2)
	stopPlanning := false
	var transcriptSummary string

loop:
	for i := 0; i < maxTurns; i++ {
		turn := i + 1
		succeededForTurn := make([]intents.Intent, 0, len(extSucceeded)+len(completed))
		succeededForTurn = append(succeededForTurn, extSucceeded...)
		succeededForTurn = append(succeededForTurn, completed...)
		failedForTurn := make([]agentcontext.FailedIntent, 0, len(extFailed)+len(failedIntents))
		failedForTurn = append(failedForTurn, extFailed...)
		failedForTurn = append(failedForTurn, failedIntents...)
		turnCtx := agentcontext.ComposeTurnContext(params, text, succeededForTurn, failedForTurn, gathered, hostCursor)
		next, err := e.agent.NextIntent(context.Background(), turnCtx)
		if err != nil {
			trace = appendTurnTrace(trace, turn, "model_error")
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		}
		if err := next.Validate(); err != nil {
			trace = appendTurnTrace(trace, turn, "invalid_intent")
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		}
		trace = appendTurnTrace(trace, turn, "intent:"+next.Summary())

		var editCtx edit.EditExecutionContext
		if next.Executable != nil {
			var planErr string
			editCtx, planErr = buildEditExecutionContext(params, next.Executable)
			if planErr != "" {
				trace = appendTurnTrace(trace, turn, "pre_execute_error")
				if maxRetries > 0 {
					trace = appendTurnTrace(trace, turn, "retry:pre_execute")
					failedIntents = append(failedIntents, agentcontext.FailedIntent{
						Intent: next,
						Phase:  agentcontext.PhasePreExecute,
						Reason: planErr,
					})
					gathered = appendGatheredNote(gathered, fmt.Sprintf("daemon rejected %q intent before execution: %s; retry with corrected intent", next.Executable.Kind, planErr))
					maxRetries--
					continue
				}
				return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
			}
		}

		if c := next.Control; c != nil && c.Kind == intents.ControlIntentKindRequestContext {
			contextRounds++
			consecutiveContextReq++
			if e.maxContextRounds > 0 && contextRounds > e.maxContextRounds {
				trace = appendTurnTrace(trace, turn, "context:cap:max_rounds")
				return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
			}
			if e.maxConsecutiveContextReq > 0 && consecutiveContextReq > e.maxConsecutiveContextReq {
				trace = appendTurnTrace(trace, turn, "context:cap:consecutive_requests")
				return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
			}
		}

		out, err := e.intentHandler.Handle(dispatch.HandleInput{
			Params:   params,
			Gathered: gathered,
			Intent:   next,
			EditCtx:  editCtx,
		})
		if err != nil {
			if next.Executable != nil {
				trace = appendTurnTrace(trace, turn, "execute_error")
				if maxRetries > 0 {
					trace = appendTurnTrace(trace, turn, "retry:execute")
					failedIntents = append(failedIntents, agentcontext.FailedIntent{
						Intent: next,
						Phase:  agentcontext.PhaseDispatch,
						Reason: err.Error(),
					})
					gathered = appendGatheredNote(gathered, fmt.Sprintf("daemon execution failed for %q intent: %v; retry with corrected intent", next.Executable.Kind, err))
					maxRetries--
					continue
				}
			} else {
				trace = appendTurnTrace(trace, turn, "context:fulfill_error")
			}
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		}

		switch {
		case out.Control != nil:
			cr := out.Control
			if cr.Done != nil {
				transcriptSummary = cr.Done.Summary
				break loop
			}
			if cr.Fulfilled != nil {
				updated := cr.Fulfilled.UpdatedGathered
				if e.maxContextBytes > 0 && agentcontext.EstimatedGatheredBytes(updated) > e.maxContextBytes {
					trace = appendTurnTrace(trace, turn, "context:cap:byte_budget")
					return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
				}
				trace = appendTurnTrace(trace, turn, "context:fulfilled")
				gathered = updated
				completed = append(completed, next)
				continue
			}
			trace = appendTurnTrace(trace, turn, "invalid_control_outcome")
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		case out.Executable != nil:
			maxRetries = e.maxIntentRetries
			if maxRetries < 0 {
				maxRetries = 0
			}
			failedIntents = nil
			consecutiveContextReq = 0
			st := out.Executable
			switch {
			case st.EditDirective != nil:
				if st.EditDirective.Kind == "success" {
					for j := range st.EditDirective.Actions {
						if st.EditDirective.Actions[j].EditId == "" {
							st.EditDirective.Actions[j].EditId = fmt.Sprintf("edit-%d", editCounter)
							editCounter++
						}
					}
				}
				directives = append(directives, protocol.VoiceTranscriptDirective{Kind: "edit", EditDirective: st.EditDirective})
				trace = appendTurnTrace(trace, turn, "result:edit:"+st.EditDirective.Kind)
				completed = append(completed, next)
				appendSourceIntentForDirective(&batchSourceIntents, next)
			case st.CommandDirective != nil:
				directives = append(directives, protocol.VoiceTranscriptDirective{Kind: "command", CommandDirective: st.CommandDirective})
				trace = appendTurnTrace(trace, turn, "result:command")
				completed = append(completed, next)
				appendSourceIntentForDirective(&batchSourceIntents, next)
			case st.NavigationDirective != nil:
				directives = append(directives, protocol.VoiceTranscriptDirective{Kind: "navigate", NavigationDirective: st.NavigationDirective})
				trace = appendTurnTrace(trace, turn, "result:navigate")
				completed = append(completed, next)
				appendSourceIntentForDirective(&batchSourceIntents, next)
			case st.UndoDirective != nil:
				directives = append(directives, protocol.VoiceTranscriptDirective{Kind: "undo", UndoDirective: st.UndoDirective})
				trace = appendTurnTrace(trace, turn, "result:undo:"+st.UndoDirective.Scope)
				completed = append(completed, next)
				appendSourceIntentForDirective(&batchSourceIntents, next)
			default:
				trace = appendTurnTrace(trace, turn, "invalid_executable_outcome")
				return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
			}
		default:
			trace = appendTurnTrace(trace, turn, "empty_handle_outcome")
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		}
		if stopPlanning {
			break
		}
	}
	if len(completed) >= maxTurns {
		trace = appendTrace(trace, "cap:max_turns")
		return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
	}
	result := protocol.VoiceTranscriptResult{
		Accepted:   true,
		Directives: directives,
		Summary:    transcriptSummary,
	}
	var pending *agentcontext.DirectiveApplyBatch
	if len(directives) > 0 {
		bid, err := newDirectiveApplyBatchID()
		if err != nil {
			trace = appendTrace(trace, "directive_apply_batch_id_error")
			return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
		}
		result.ApplyBatchId = bid
		pending = &agentcontext.DirectiveApplyBatch{
			ID:            bid,
			SourceIntents: append([]intents.Intent(nil), batchSourceIntents...),
		}
	}
	if err := result.Validate(); err != nil {
		return protocol.VoiceTranscriptResult{Accepted: false}, gathered, nil, true
	}
	return result, gathered, pending, true
}
