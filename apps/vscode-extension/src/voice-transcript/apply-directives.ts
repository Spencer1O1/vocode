import type { VoiceTranscriptDirective } from "@vocode/protocol";

import { dispatchTranscript } from "../directives/dispatch";
import {
  beginTranscriptUndoSession,
  finalizeTranscriptUndoSessionIfEditsApplied,
} from "../directives/undo/transcript-undo-ledger";
import type { CommandApplyUiHandlers, TranscriptApplyContext } from "./context";

export type DirectiveApplyOutcome = {
  status: "ok" | "failed" | "skipped";
  message?: string;
};

export type ApplyTranscriptResult = {
  outcomes: DirectiveApplyOutcome[];
  /** Paths of files modified but not yet saved (only set when previewMode is true). */
  previewPaths: string[];
};

/**
 * Applies a daemon directive batch to the workspace (edits, commands, navigation, undo).
 * Used for every `voice.transcript` path (voice commits and command-palette send).
 * Returns one outcome per directive (stops after the first failure).
 *
 * When `options.previewMode` is true, edits are applied but not saved. The returned
 * `previewPaths` lists the files that were modified and are awaiting accept/reject.
 */
export type ApplyTranscriptProgressEvent = {
  index: number;
  phase: "start" | "complete";
  outcome?: DirectiveApplyOutcome;
};

export async function applyDirectives(
  directives: readonly VoiceTranscriptDirective[],
  activeDocumentPath: string,
  options?: {
    onProgress?: (event: ApplyTranscriptProgressEvent) => void;
    /** When set, shell command stdout/stderr is streamed to the main panel for this pending row. */
    commandApplyUi?: CommandApplyUiHandlers;
    previewMode?: boolean;
  },
): Promise<ApplyTranscriptResult> {
  const ctx: TranscriptApplyContext = {
    activeDocumentPath,
    editLocations: {},
    ...(options?.commandApplyUi !== undefined
      ? { commandApplyUi: options.commandApplyUi }
      : {}),
    previewMode: options?.previewMode,
  };
  const previewPaths: string[] = [];

  const dirs = directives ?? [];
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
        return { outcomes, previewPaths };
      }
      if (dispatchOutcome.editedPaths) {
        previewPaths.push(...dispatchOutcome.editedPaths);
      }
      const ok: DirectiveApplyOutcome = { status: "ok" };
      outcomes.push(ok);
      options?.onProgress?.({ index: i, phase: "complete", outcome: ok });
    }
  } finally {
    finalizeTranscriptUndoSessionIfEditsApplied();
  }
  return { outcomes, previewPaths };
}
