import type { RenameDirective } from "@vocode/protocol";
import * as vscode from "vscode";

import type { DirectiveDispatchOutcome } from "../dispatch";

export async function dispatchRename(
  rename: RenameDirective | undefined,
): Promise<DirectiveDispatchOutcome> {
  if (!rename) {
    return { ok: false, message: "missing rename directive" };
  }
  const { path, position, newName } = rename;
  if (!path || !newName) {
    return { ok: false, message: "rename directive missing path or newName" };
  }
  try {
    const uri = vscode.Uri.file(path);
    const doc = await vscode.workspace.openTextDocument(uri);
    const editor = await vscode.window.showTextDocument(doc, {
      preview: false,
    });
    const pos = new vscode.Position(position.line, position.character);
    editor.selection = new vscode.Selection(pos, pos);
    editor.revealRange(new vscode.Range(pos, pos));

    const edit = (await vscode.commands.executeCommand(
      "vscode.executeDocumentRenameProvider",
      uri,
      pos,
      newName,
    )) as vscode.WorkspaceEdit | undefined;

    if (!edit) {
      return { ok: false, message: "rename failed: no WorkspaceEdit returned" };
    }
    const applied = await vscode.workspace.applyEdit(edit);
    if (!applied) {
      return {
        ok: false,
        message: "rename failed: workspace.applyEdit returned false",
      };
    }

    // Best-effort save all dirty editors after rename.
    await vscode.workspace.saveAll(false);
    return { ok: true };
  } catch (err) {
    const message =
      err instanceof Error ? err.message : "rename dispatch failed";
    return { ok: false, message };
  }
}
