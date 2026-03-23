import type {
  EditApplyParams,
  ReplaceBetweenAnchorsAction,
} from "@vocode/protocol";
import { isEditApplyResult } from "@vocode/protocol";
import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

function applyReplaceBetweenAnchors(
  documentText: string,
  action: ReplaceBetweenAnchorsAction,
): string {
  const beforeIndex = documentText.indexOf(action.anchor.before);
  if (beforeIndex === -1) {
    throw new Error(
      `Could not find before anchor: ${JSON.stringify(action.anchor.before)}`,
    );
  }

  const searchStart = beforeIndex + action.anchor.before.length;
  const afterIndex = documentText.indexOf(action.anchor.after, searchStart);
  if (afterIndex === -1) {
    throw new Error(
      `Could not find after anchor: ${JSON.stringify(action.anchor.after)}`,
    );
  }

  const prefix = documentText.slice(0, searchStart);
  const suffix = documentText.slice(afterIndex);

  return `${prefix}${action.newText}${suffix}`;
}

async function replaceWholeDocument(
  editor: vscode.TextEditor,
  newText: string,
): Promise<void> {
  const document = editor.document;
  const lastLine = document.lineAt(document.lineCount - 1);
  const fullRange = new vscode.Range(
    new vscode.Position(0, 0),
    lastLine.rangeIncludingLineBreak.end,
  );

  const success = await editor.edit((editBuilder) => {
    editBuilder.replace(fullRange, newText);
  });

  if (!success) {
    throw new Error("VS Code failed to apply the edit.");
  }
}

export const applyEditCommand: CommandDefinition = {
  id: "vocode.applyEdit",
  requiresDaemon: true,
  run: async (client) => {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      void vscode.window.showErrorMessage("No active editor.");
      return;
    }

    const instruction = await vscode.window.showInputBox({
      title: "Vocode Apply Edit",
      prompt: "Describe the edit to apply",
      placeHolder: "Insert a console.log inside the current function",
      ignoreFocusOut: true,
    });

    if (!instruction) {
      return;
    }

    const document = editor.document;
    const params: EditApplyParams = {
      instruction,
      activeFile: document.uri.fsPath,
      fileText: document.getText(),
    };

    const result = await client.applyEdit(params);

    if (!isEditApplyResult(result)) {
      throw new Error("Daemon returned an invalid edit/apply result.");
    }

    let nextText = document.getText();

    for (const action of result.actions) {
      if (action.path !== document.uri.fsPath) {
        throw new Error(`Received action for unsupported file: ${action.path}`);
      }

      for (const action of result.actions) {
        switch (action.kind) {
          case "replace_between_anchors":
            nextText = applyReplaceBetweenAnchors(nextText, action);
            break;
          default:
            throw new Error(`Unsupported action kind: ${action.kind}`);
        }
      }
    }

    await replaceWholeDocument(editor, nextText);
    void vscode.window.showInformationMessage("Vocode edit applied.");
  },
};
