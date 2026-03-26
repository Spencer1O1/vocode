import * as vscode from "vscode";

import type { CommandDefinition } from "./types";

export const startVoiceCommand: CommandDefinition = {
  id: "vocode.startVoice",
  requiresDaemon: true,
  run: (_client, services) => {
    if (services.voiceSession.isRunning()) {
      void vscode.window.showInformationMessage("Vocode is already listening.");
      return;
    }

    try {
      if (!services.voiceSidecar) {
        throw new Error("voice sidecar is not running");
      }
      services.voiceSidecar.start();
      services.voiceSession.start();
      services.voiceStatus.setListening();
      void vscode.window.showInformationMessage("Vocode started listening.");
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Unknown microphone error";
      void vscode.window.showWarningMessage(
        `Unable to start microphone capture: ${message}`,
      );
      services.voiceSession.stop();
      services.voiceStatus.setIdle();
    }
  },
};
