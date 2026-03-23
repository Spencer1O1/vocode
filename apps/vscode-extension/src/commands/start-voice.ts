import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

export const startVoiceCommand: CommandDefinition = {
  id: "vocode.startVoice",
  requiresDaemon: true,
  run: () => {
    void vscode.window.showInformationMessage("Vocode: Start Voice");
  },
};
