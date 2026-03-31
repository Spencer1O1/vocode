import type { FormatDirective } from "@vocode/protocol";
import * as vscode from "vscode";

import type { DirectiveDispatchOutcome } from "../dispatch";

export async function dispatchFormat(
  directive: FormatDirective | undefined,
): Promise<DirectiveDispatchOutcome> {
  if (!directive) return { ok: false, message: "missing formatDirective" };
  try {
    const uri = vscode.Uri.file(directive.path);
    const doc = await vscode.workspace.openTextDocument(uri);
    await vscode.window.showTextDocument(doc, { preview: false });

    let edits: vscode.TextEdit[] | undefined;
    if (directive.scope === "selection") {
      const r = directive.range;
      if (!r) {
        return { ok: false, message: "format selection requires range" };
      }
      const range = new vscode.Range(
        new vscode.Position(r.startLine, r.startChar),
        new vscode.Position(r.endLine, r.endChar),
      );
      edits = (await vscode.commands.executeCommand(
        "vscode.executeFormatRangeProvider",
        uri,
        range,
        { insertSpaces: true, tabSize: 2 },
      )) as vscode.TextEdit[] | undefined;
    } else {
      edits = (await vscode.commands.executeCommand(
        "vscode.executeFormatDocumentProvider",
        uri,
        { insertSpaces: true, tabSize: 2 },
      )) as vscode.TextEdit[] | undefined;
    }

    if (!edits || edits.length === 0) {
      return { ok: false, message: "format failed: no edits returned" };
    }

    const wsEdit = new vscode.WorkspaceEdit();
    for (const e of edits) {
      wsEdit.replace(uri, e.range, e.newText);
    }
    const applied = await vscode.workspace.applyEdit(wsEdit);
    if (!applied) {
      return {
        ok: false,
        message: "format failed: workspace.applyEdit returned false",
      };
    }
    await vscode.workspace.saveAll(false);
    return { ok: true };
  } catch (err) {
    const message =
      err instanceof Error ? err.message : "format dispatch failed";
    return { ok: false, message };
  }
}
