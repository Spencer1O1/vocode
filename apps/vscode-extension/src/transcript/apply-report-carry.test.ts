import assert from "node:assert/strict";
import { beforeEach, test } from "node:test";

import type {
  VoiceTranscriptParams,
  VoiceTranscriptResult,
} from "@vocode/protocol";

import {
  clearPendingApplyReport,
  mergeCarriedApplyReportParams,
  carryApplyReportIntoNextRpc,
} from "./apply-report-carry";

beforeEach(() => {
  clearPendingApplyReport();
});

function commandDirective(n: number) {
  return {
    kind: "command" as const,
    commandDirective: { command: "echo", args: [String(n)] },
  };
}

function sevenDirectiveResult(batchId: string): VoiceTranscriptResult {
  return {
    success: true,
    applyBatchId: batchId,
    directives: [
      commandDirective(0),
      commandDirective(1),
      commandDirective(2),
      commandDirective(3),
      commandDirective(4),
      commandDirective(5),
      commandDirective(6),
    ],
  };
}

test("mergeCarriedApplyReportParams leaves base unchanged when nothing carried", () => {
  const base: VoiceTranscriptParams = {
    text: "hi",
    contextSessionId: "s1",
  };
  const out = mergeCarriedApplyReportParams(base);
  assert.deepEqual(out, base);
  assert.equal(out.reportApplyBatchId, undefined);
  assert.equal(out.lastBatchApply, undefined);
});

test("carryApplyReportIntoNextRpc does not carry when outcomes length mismatches directives", () => {
  carryApplyReportIntoNextRpc(sevenDirectiveResult("b1"), [
    { status: "ok" },
    { status: "ok" },
  ]);
  const out = mergeCarriedApplyReportParams({ text: "x", contextSessionId: "s" });
  assert.equal(out.reportApplyBatchId, undefined);
  assert.equal(out.lastBatchApply, undefined);
});

test("seven directives: fail at index 3, tail skipped — next params carry batch id and statuses", () => {
  const batchId = "repair-batch-7";
  const result = sevenDirectiveResult(batchId);
  const outcomes = [
    { status: "ok" as const },
    { status: "ok" as const },
    { status: "ok" as const },
    { status: "failed" as const, message: "boom" },
    { status: "skipped" as const, message: "not attempted" },
    { status: "skipped" as const, message: "not attempted" },
    { status: "skipped" as const, message: "not attempted" },
  ];
  carryApplyReportIntoNextRpc(result, outcomes);

  const merged = mergeCarriedApplyReportParams({
    text: "fix it",
    contextSessionId: "sess",
  });

  assert.equal(merged.text, "fix it");
  assert.equal(merged.contextSessionId, "sess");
  assert.equal(merged.reportApplyBatchId, batchId);
  assert.ok(merged.lastBatchApply);
  assert.equal(merged.lastBatchApply.length, 7);

  for (let i = 0; i < 3; i++) {
    assert.equal(merged.lastBatchApply[i].status, "ok");
    assert.equal(merged.lastBatchApply[i].message, undefined);
  }
  assert.equal(merged.lastBatchApply[3].status, "failed");
  assert.equal(merged.lastBatchApply[3].message, "boom");
  for (let i = 4; i < 7; i++) {
    assert.equal(merged.lastBatchApply[i].status, "skipped");
    assert.equal(merged.lastBatchApply[i].message, "not attempted");
  }

  const again = mergeCarriedApplyReportParams({
    text: "second",
    contextSessionId: "sess",
  });
  assert.equal(again.reportApplyBatchId, undefined);
  assert.equal(again.lastBatchApply, undefined);
});

test("carryApplyReportIntoNextRpc ignores failed transcript result", () => {
  carryApplyReportIntoNextRpc(
    { success: false, directives: [commandDirective(0)], applyBatchId: "x" },
    [{ status: "ok" }],
  );
  const out = mergeCarriedApplyReportParams({
    text: "t",
    contextSessionId: "s",
  });
  assert.equal(out.reportApplyBatchId, undefined);
});

