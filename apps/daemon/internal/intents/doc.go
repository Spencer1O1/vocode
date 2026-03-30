// Package intents defines executable host steps validated by the daemon and routed by dispatch.
//
// # Taxonomy
//
// [Intent] is always executable (edit, command, navigate, undo): one top-level "kind" and payload.
//
// Turn-level outcomes that are not host directives — irrelevant, finish ("done"), gather context
// ("request_context" with [agentcontext.GatherContextSpec]) — live on the daemon agent turn result
// type in package agent, not as [Intent]. Finish summary validation is [agent.ValidateFinishSummary].
//
// Payload types on [Intent]:
//   - [EditIntent]
//   - [CommandIntent]
//   - [NavigationIntent]
//   - [UndoIntent]
package intents
