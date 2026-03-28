import * as vscode from "vscode";

import type { CommandDefinition } from "./types";
import { runUndoLastEditWithDeps } from "./undo-last-edit-logic";

export const undoLastEditCommand: CommandDefinition = {
  id: "vocode.undoLastEdit",
  requiresDaemon: false,
  run: async () => {
    await runUndoLastEditWithDeps({
      hasActiveEditor: () => Boolean(vscode.window.activeTextEditor),
      executeUndo: () => vscode.commands.executeCommand("undo"),
      showWarning: (message) => {
        void vscode.window.showWarningMessage(message);
      },
    });
  },
};
