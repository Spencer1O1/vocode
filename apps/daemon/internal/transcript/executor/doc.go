// Package executor implements the narrow-model voice.transcript pipeline: transcript classification
// (prompt/classifier), scope intent (prompt/scope_intent), then scoped edit or other deterministic
// directive builders (format, rename, search, etc.). The daemon applies host directives in a single
// shot per utterance—no iterative intent/repair loop inside this package.
//
// Entry: executor.go, helpers.go.
package executor
