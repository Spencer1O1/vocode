// Package agentcontext holds the structured input the iterative [agent.ModelClient] sees each turn
// (one voice.transcript / one NextIntent turn): transcript, editor snapshot, gathered extras, intent history.
//
// It lives beside package [agent] so the boundary is clear: [agent] runs the model client; agentcontext
// is the value passed into [agent.Agent.NextIntent]. The name avoids a top-level type called just
// "Context", which collides mentally with [context.Context].
//
// Turn shape:
//   - [TurnContext.TranscriptText]: user utterance for this RPC (stable across turns in one Execute).
//   - [TurnContext.SucceededIntents]: intents from the last [DirectiveApplyBatch] the host reported as ok (via lastBatchApply + reportApplyBatchId),
//     plus intents dispatched in the current Execute (executable → directive, or request_context → fulfilled).
//   - [TurnContext.FailedIntents]: pre-execute, dispatch, or extension apply failures ([PhaseExtension]).
//   - [TurnContext.Editor]: active path and caret symbol from this RPC’s params (host refreshes each transcript).
//   - [TurnContext.Gathered]: excerpts (active file + request_context), symbols, notes; retained in the daemon
//     between transcripts when the host sends the same contextSessionId ([VoiceSessionStore]).
//
// The extension sends cursorPosition; the daemon resolves [EditorSnapshot.CursorSymbol] via symbols/tags.
package agentcontext
