import type { UndoDirective } from "@vocode/protocol";

import type { DirectiveDispatchOutcome } from "../dispatch";
import { applyUndoDirective } from "./transcript-undo-ledger";

/** Applies one undo directive (host undo stack / transcript ledger). */
export function dispatchUndo(
  directive: UndoDirective | undefined,
): Promise<DirectiveDispatchOutcome> {
  return applyUndoDirective(directive)
    .then((ok) => {
      if (ok) return { ok: true };
      return { ok: false, message: "undo directive not applied" };
    })
    .catch((err) => {
      const message =
        err instanceof Error ? err.message : "undo dispatch failed";
      return { ok: false, message };
    });
}
