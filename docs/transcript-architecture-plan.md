# Transcript / gathered context architecture (implemented plan)

## Goals

- **Voice-first UX:** No user-facing sessions or clear-context UX; internal keys and resets only.
- **Thinner wire:** core sends directive batches via `host.applyDirectives`; host returns per-directive `{status, message}[]` only.
- **Batching:** Multiple directives per `voice.transcript` remain allowed; extension stops on first failure and reports **one row per directive** (tail = skipped / not attempted).
- **Long mic on:** **Idle reset** clears stored voice session for a context key; **rolling cap** trims gathered excerpts while **never dropping the current `activeFile` excerpt**.

## Phases

### A — Wire + pending apply batch (core authority)

- Core→host: `host.applyDirectives` with `applyBatchId`, `activeFile`, `directives[]`.
- Host→core: `HostApplyResult.items[]` (`status`: `ok` | `failed` | `skipped`, optional `message`).
- `voice.transcript` returns `VoiceTranscriptCompletion` (classification + UI disposition + optional search/Q&A payloads). It does not return directives on the completion object; directives are applied via `host.applyDirectives` in the same RPC when needed.
- `transcript.Service` (`apps/core/internal/transcript`) holds `session.VoiceSessionStore`, `executeMu`, and runs `pipeline.Execute` per utterance under the mutex. **Single-shot:** one pipeline pass and at most one host apply batch per utterance (no core-side repair loop).
- Session state lives in `apps/core/internal/transcript/session`; clarify/selection/file-selection overlays and flow routing live in `apps/core/internal/transcript/pipeline`, `apps/core/internal/transcript/clarify`, and `apps/core/internal/flows`. Per-RPC caps via `daemonConfig` on `voice.transcript` (`sessionIdleResetMs`, `maxGatheredBytes`, `maxGatheredExcerpts`; RPC field name is historical).

### B — Gathered / excerpt policy

- Rolling caps and idle reset are applied in the core transcript + session layers (see `transcript/idle`, `session` types, and pipeline). Never drop the active file excerpt when trimming.

### C — Control RPC

- `controlRequest`: `cancel_clarify` | `cancel_selection` (close code match list / selection session without spoken text).

## Extension

- Code layout: `apps/vscode-extension/src/voice-transcript/` (RPC helpers, `apply-directives`, workspace root), `src/ui/panel/` (main webview provider + store).
- Apply directives when the core requests them via `host.applyDirectives` during `voice.transcript`; return per-directive outcomes in one shot.
- Committed transcript handlers are serialized so a new user transcript does not start until the current RPC finishes.
- Failure messages come from directive dispatchers and are surfaced in `HostApplyResult.items[i].message`.
