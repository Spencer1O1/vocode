import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

export const runCommand: CommandDefinition = {
  id: "vocode.runCommand",
  requiresDaemon: false,
  run: () => {
    void vscode.window.showInformationMessage("Vocode: Run Command");
  },
};
