// Package transcript wires voice.transcript on the daemon: RPC service, optional coalescing queue,
// session / directive-batch correlation, and delegation to [executor.Executor] for the narrow-model pipeline.
//
// Subpackages:
//   - [vocoding.net/vocode/v2/apps/daemon/internal/transcript/executor] — classifier / scope / scoped edit → directives
//   - [vocoding.net/vocode/v2/apps/daemon/internal/transcript/voicesession] — load/save voice session + apply report intake
//   - [vocoding.net/vocode/v2/apps/daemon/internal/transcript/config] — env for transcript path
//
// Root files: service.go, service_session.go, service_worker.go.
package transcript
