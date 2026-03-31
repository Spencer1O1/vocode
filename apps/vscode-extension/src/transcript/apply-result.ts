import type { VoiceTranscriptResult } from "@vocode/protocol";

import { dispatchTranscript } from "../directives/dispatch";
import {
  beginTranscriptUndoSession,
  finalizeTranscriptUndoSessionIfEditsApplied,
} from "../directives/undo/transcript-undo-ledger";
import type { TranscriptApplyContext } from "./context";

export type DirectiveApplyOutcome = {
  status: "ok" | "failed" | "skipped";
  message?: string;
};

/**
 * Applies a daemon `VoiceTranscriptResult` to the workspace (edits, commands, navigation, undo).
 * Used by voice and by the manual “send transcript” command — not command-specific.
 * Returns one outcome per directive (stops after the first failure).
 */
export type ApplyTranscriptProgressEvent = {
  index: number;
  phase: "start" | "complete";
  outcome?: DirectiveApplyOutcome;
};

export async function applyTranscriptResult(
  result: VoiceTranscriptResult,
  activeDocumentPath: string,
  options?: {
    onProgress?: (event: ApplyTranscriptProgressEvent) => void;
  },
): Promise<DirectiveApplyOutcome[]> {
  if (!result.success) {
    return [];
  }

  const ctx: TranscriptApplyContext = {
    activeDocumentPath,
    editLocations: {},
  };

  const dirs = result.directives ?? [];
  const outcomes: DirectiveApplyOutcome[] = [];
  beginTranscriptUndoSession();
  try {
    for (let i = 0; i < dirs.length; i++) {
      const directive = dirs[i];
      options?.onProgress?.({ index: i, phase: "start" });
      const dispatchOutcome = await dispatchTranscript(directive, ctx);
      if (!dispatchOutcome.ok) {
        const failed: DirectiveApplyOutcome = {
          status: "failed",
          message: dispatchOutcome.message ?? "Directive failed to apply.",
        };
        outcomes.push(failed);
        options?.onProgress?.({ index: i, phase: "complete", outcome: failed });
        for (let j = i + 1; j < dirs.length; j++) {
          const skipped: DirectiveApplyOutcome = {
            status: "skipped",
            message: "not attempted",
          };
          outcomes.push(skipped);
          options?.onProgress?.({
            index: j,
            phase: "complete",
            outcome: skipped,
          });
        }
        return outcomes;
      }
      const ok: DirectiveApplyOutcome = { status: "ok" };
      outcomes.push(ok);
      options?.onProgress?.({ index: i, phase: "complete", outcome: ok });
    }
  } finally {
    finalizeTranscriptUndoSessionIfEditsApplied();
  }
  return outcomes;
}
