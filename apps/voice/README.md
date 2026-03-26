# Voice Sidecar (`apps/voice`)

`apps/voice` is a dedicated process for voice I/O concerns (microphone capture
and speech-to-text orchestration), intentionally separate from:

- `apps/daemon` (planning + semantic policy + action-plan dispatch)
- `apps/vscode-extension` (UI + process orchestration + editor mechanics)

## Purpose

This sidecar is the place to implement:

- cross-platform microphone capture
- audio buffering/chunking
- STT integrations (cloud/local)
- transcript event emission back to the extension

It should not contain planning/action-plan logic.

## Transport

The initial skeleton uses JSON lines over stdio:

- Extension writes requests to sidecar stdin.
- Sidecar writes events/responses to stdout.

Current request/event shapes are defined in `internal/app`.

## Binary

The sidecar command entrypoint is:

- `apps/voice/cmd/vocode-voiced`
