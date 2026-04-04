import type { CodeActionDirective } from "@vocode/protocol";
import * as vscode from "vscode";

import type { DirectiveDispatchOutcome } from "../dispatch";

function toRange(
  r: CodeActionDirective["range"] | undefined,
): vscode.Range | undefined {
  if (!r) return undefined;
  return new vscode.Range(
    new vscode.Position(r.startLine, r.startChar),
    new vscode.Position(r.endLine, r.endChar),
  );
}

export async function dispatchCodeAction(
  directive: CodeActionDirective | undefined,
): Promise<DirectiveDispatchOutcome> {
  if (!directive) {
    return { ok: false, message: "missing codeActionDirective" };
  }
  try {
    const uri = vscode.Uri.file(directive.path);
    const doc = await vscode.workspace.openTextDocument(uri);
    const editor = await vscode.window.showTextDocument(doc, {
      preview: false,
    });

    const range = toRange(directive.range) ?? editor.selection;
    const kind = directive.actionKind;

    const actions = (await vscode.commands.executeCommand(
      "vscode.executeCodeActionProvider",
      uri,
      range,
      kind,
    )) as (vscode.CodeAction | vscode.Command)[] | undefined;

    if (!actions || actions.length === 0) {
      if (directive.actionKind === "source.organizeImports") {
        return { ok: true };
      }
      return {
        ok: false,
        message: `code_action failed: no actions for ${directive.actionKind}`,
      };
    }

    const preferred = directive.preferredTitleIncludes?.trim().toLowerCase();
    const pick = actions.find((a) => {
      if ("title" in a && preferred) {
        return a.title.trim().toLowerCase().includes(preferred);
      }
      return true;
    });

    if (!pick) {
      return {
        ok: false,
        message: "code_action failed: no matching action title",
      };
    }

    if ("edit" in pick && pick.edit) {
      const applied = await vscode.workspace.applyEdit(pick.edit);
      if (!applied) {
        return {
          ok: false,
          message: "code_action failed: workspace.applyEdit returned false",
        };
      }
    }

    if ("command" in pick && pick.command) {
      if (typeof pick.command === "string") {
        await vscode.commands.executeCommand(pick.command);
      } else {
        await vscode.commands.executeCommand(
          pick.command.command,
          ...(pick.command.arguments ?? []),
        );
      }
    }

    // Save changes.
    await vscode.workspace.saveAll(false);
    return { ok: true };
  } catch (err) {
    const message =
      err instanceof Error ? err.message : "code_action dispatch failed";
    return { ok: false, message };
  }
}
