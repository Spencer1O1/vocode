import * as vscode from "vscode";

import { DaemonClient } from "./client/daemon-client";
import { registerAllCommands } from "./commands";
import {
  type ExtensionServices,
  VoiceSessionController,
} from "./commands/services";
import { spawnDaemon } from "./daemon/spawn";
import { VoiceStatusIndicator } from "./ui/status-bar";
import { MicrophoneCapture } from "./voice/microphone";
import { VoiceSidecarClient } from "./voice-sidecar/client";
import { spawnVoiceSidecar } from "./voice-sidecar/spawn";

function createServices(
  context: vscode.ExtensionContext,
  voiceStatus: VoiceStatusIndicator,
): ExtensionServices {
  const microphone = new MicrophoneCapture();
  const debugAudioLogging = vscode.workspace
    .getConfiguration("vocode")
    .get<boolean>("debugAudioLogging", false);

  const audioChunkSubscription = microphone.onAudioChunk(({ data }) => {
    if (debugAudioLogging) {
      console.debug(
        `Vocode microphone chunk captured (${data.byteLength} bytes)`,
      );
    }
  });

  context.subscriptions.push(microphone, audioChunkSubscription);
  try {
    const daemon = spawnDaemon(context);
    console.log(`Vocode daemon started from ${daemon.binaryPath}`);

    const voice = spawnVoiceSidecar(context);
    console.log(`Vocode voice sidecar started from ${voice.binaryPath}`);

    return {
      client: new DaemonClient(daemon.process),
      voiceStatus,
      voiceSession: new VoiceSessionController(),
      microphone,
      voiceSidecar: new VoiceSidecarClient(voice.process),
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Unknown daemon startup error";

    console.error(message);
    void vscode.window.showErrorMessage(
      `Failed to start Vocode daemon: ${message}`,
    );

    return {
      client: null,
      voiceStatus,
      voiceSession: new VoiceSessionController(),
      microphone,
      voiceSidecar: null,
    };
  }
}

export function activate(context: vscode.ExtensionContext) {
  console.log("Vocode extension activated");

  const voiceStatus = new VoiceStatusIndicator();
  const services = createServices(context, voiceStatus);

  context.subscriptions.push(voiceStatus, ...registerAllCommands(services), {
    dispose: () => {
      services.client?.dispose();
    },
  });
}

export function deactivate() {}
