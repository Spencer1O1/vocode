import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

export const stopVoiceCommand: CommandDefinition = {
  id: "vocode.stopVoice",
  requiresDaemon: true,
  run: () => {
    void vscode.window.showInformationMessage("Vocode: Stop Voice");
  },
};
