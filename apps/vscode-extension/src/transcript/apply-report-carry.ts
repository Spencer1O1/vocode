import type {
  VoiceTranscriptDirectiveApplyItem,
  VoiceTranscriptParams,
  VoiceTranscriptResult,
} from "@vocode/protocol";

export type DirectiveApplyOutcome = {
  status: "ok" | "failed" | "skipped";
  message?: string;
};

// Single-slot "carry": values produced by applying the current directives are
// attached to the *next* `voice.transcript` RPC.
let carriedReportApplyBatchId: string | undefined;
let carriedLastBatchApply: VoiceTranscriptDirectiveApplyItem[] | undefined;

/** Drops any carried apply report (e.g. voice session ended before next RPC). */
export function clearPendingApplyReport(): void {
  carriedReportApplyBatchId = undefined;
  carriedLastBatchApply = undefined;
}

/** Merges carried `lastBatchApply` + `reportApplyBatchId` into next RPC params. */
export function mergeCarriedApplyReportParams(
  base: VoiceTranscriptParams,
): VoiceTranscriptParams {
  const out: VoiceTranscriptParams = { ...base };
  if (carriedReportApplyBatchId !== undefined) {
    out.reportApplyBatchId = carriedReportApplyBatchId;
    carriedReportApplyBatchId = undefined;
  }
  if (carriedLastBatchApply !== undefined) {
    out.lastBatchApply = carriedLastBatchApply;
    carriedLastBatchApply = undefined;
  }
  return out;
}

/** After applying directives, carry apply report fields into the next RPC. */
export function carryApplyReportIntoNextRpc(
  result: VoiceTranscriptResult,
  outcomes: DirectiveApplyOutcome[],
): void {
  const dirs = result.directives ?? [];
  const batchId = result.applyBatchId?.trim() ?? "";
  if (
    result.success &&
    dirs.length > 0 &&
    batchId !== "" &&
    outcomes.length === dirs.length
  ) {
    carriedReportApplyBatchId = batchId;
    carriedLastBatchApply = outcomes.map((o) => ({
      status: o.status,
      ...(o.message !== undefined && o.message !== ""
        ? { message: o.message }
        : {}),
    }));
  }
}

