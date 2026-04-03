# gRPC + Protobuf Duplex Backlog

## Why
The current “best UX” duplex implementation uses line-delimited JSON-RPC over stdio:
- core (`vocode-cored`) plans and sends directive batches to the extension (`host.applyDirectives`)
- extension executes and returns per-directive outcomes
- core continues planning until directives are empty

This works, but the protocol boundary would be cleaner with a proper streaming transport.

## Target streaming shape
Implement a duplex gRPC stream (one stream per committed transcript / `contextSessionId` key):

1. Core->Host messages: `DirectiveBatch`
   - `applyBatchId`
   - `activeFile`
   - `directives[]`
2. Host->Core messages: `ApplyOutcomeBatch`
   - `applyBatchId`
   - `items[]` with `status: ok|failed|skipped` and optional `message`
3. Core completes by sending a final `VoiceTranscriptCompletion` (with `success=true` plus classification fields like `transcriptOutcome` and `uiDisposition`).

## Mapping from current JSON-RPC
- `host.applyDirectives` request.params becomes `DirectiveBatch`
- `host.applyDirectives` result becomes `ApplyOutcomeBatch.items`
- Each committed utterance still maps to **one** core planning pass and **at most one** host apply batch; a duplex stream would carry those batches in order, not re-run a core-side retry loop per utterance.

## Notes / constraints
- Preserve host feedback (per-directive outcomes) so the product can decide follow-up policy in the extension or a future multi-turn planner; the current `voice.transcript` path remains single-shot per RPC.

