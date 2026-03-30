import type { CommandDirective } from "@vocode/protocol";
import * as vscode from "vscode";

import type { DirectiveDispatchOutcome } from "../dispatch";
import { runAllowedCommand } from "./execute-command";

/** Runs one allowed command directive (extension executes; daemon validated shape). */
export async function dispatchCommand(
  params: CommandDirective | undefined,
): Promise<DirectiveDispatchOutcome> {
  if (!params) {
    return { ok: false, message: "missing command directive" };
  }
  const outcome = await runAllowedCommand(params);
  if (!outcome.ok) {
    const stderr = outcome.stderr.trim();
    return {
      ok: false,
      message: stderr
        ? `${(outcome.message ?? "command failed").trim()}: ${stderr}`
        : (outcome.message?.trim() ??
          "command exited non-zero or failed to run"),
    };
  }
  const line = outcome.stdout.trim();
  if (line.length > 0) {
    void vscode.window.showInformationMessage(`Vocode: ${line}`);
  }
  return { ok: true };
}
