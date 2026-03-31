import assert from "node:assert/strict";
import test from "node:test";

import { isVoiceTranscriptResult } from "./validators";

test("isVoiceTranscriptResult accepts success=true shape", () => {
  assert.equal(isVoiceTranscriptResult({ success: true }), true);
});

test("isVoiceTranscriptResult accepts summary when success", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: true,
      summary: "Renamed the handler and fixed imports.",
    }),
    true,
  );
});

test("isVoiceTranscriptResult accepts transcriptOutcome when success", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: true,
      summary: "Not a coding request.",
      transcriptOutcome: "irrelevant",
    }),
    true,
  );
});

test("isVoiceTranscriptResult rejects transcriptOutcome when not success", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: false,
      transcriptOutcome: "irrelevant",
    }),
    false,
  );
});

test("isVoiceTranscriptResult rejects summary when not success", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: false,
      summary: "oops",
    }),
    false,
  );
});

test("isVoiceTranscriptResult allows success=false minimal shape", () => {
  assert.equal(isVoiceTranscriptResult({ success: false }), true);
});

test("isVoiceTranscriptResult rejects extra keys", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: true,
      extra: 123,
    }),
    false,
  );
});

test("isVoiceTranscriptResult rejects extra keys (unexpected property)", () => {
  assert.equal(
    isVoiceTranscriptResult({
      success: true,
      unexpected: "bad",
    }),
    false,
  );
});
