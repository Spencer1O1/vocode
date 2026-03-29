import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

export const restartBackendCommand: CommandDefinition = {
  id: "vocode.restartBackend",
  requiresDaemon: false,
  run: async (services) => {
    if (!services.restartVocode) {
      void vscode.window.showErrorMessage(
        "Vocode restart is not available in this session.",
      );
      return;
    }
    await services.restartVocode();
  },
};
