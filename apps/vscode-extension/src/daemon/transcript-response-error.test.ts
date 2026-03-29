import assert from "node:assert/strict";
import test from "node:test";

import { voiceTranscriptResponseValidationError } from "./transcript-response-error";

test("detects legacy accepted:boolean transcript shape", () => {
  const message = voiceTranscriptResponseValidationError({ accepted: false });
  assert.match(message, /protocol mismatch/i);
  assert.match(message, /pnpm codegen/);
});

test("falls back to generic message for unknown shape", () => {
  const message = voiceTranscriptResponseValidationError({ ok: true });
  assert.equal(message, "Invalid voice.transcript response from daemon.");
});
