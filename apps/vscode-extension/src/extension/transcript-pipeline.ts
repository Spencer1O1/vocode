import * as vscode from "vscode";

import type { ExtensionServices } from "../commands/services";
import { FAILED_TO_PROCESS_TRANSCRIPT } from "../transcript/messages";
import { runRepairChainQueued } from "../transcript/repair-chain";
import { transcriptWorkspaceRoot } from "../transcript/workspace-root";

/**
 * Binds voice sidecar events and transcript → daemon → apply flow to the given
 * client/sidecar pair. Call again only after replacing `services.client` and
 * `services.voiceSidecar` (e.g. backend restart).
 */
export function attachTranscriptPipeline(services: ExtensionServices): void {
  const { client, voiceSidecar, voiceSession, voiceStatus, mainPanelStore } =
    services;
  if (!client || !voiceSidecar) {
    return;
  }

  let inFlightTranscripts = 0;

  voiceSidecar.onAudioMeter((evt) => {
    mainPanelStore.setAudioMeter(evt.speaking, evt.rms);
  });

  voiceSidecar.onError((evt) => {
    const message =
      typeof evt.message === "string" ? evt.message : "unknown error";
    mainPanelStore.setVoiceListening(false);
    voiceStatus.setIdle();
    if (voiceSession.isRunning()) {
      voiceSession.stop();
      void vscode.window.showWarningMessage(
        `Vocode voice sidecar error: ${message}`,
      );
    }
    voiceSidecar.stop();
  });

  voiceSidecar.onState((evt) => {
    if (evt.state !== "stopped" && evt.state !== "shutdown") {
      return;
    }
    if (!voiceSession.isRunning()) {
      return;
    }
    voiceSession.stop();
    voiceStatus.setIdle();
    mainPanelStore.setVoiceListening(false);
  });

  voiceSidecar.onTranscript((evt) => {
    if (evt.committed !== true) {
      mainPanelStore.onPartial(evt.text);
      if (!voiceSession.isRunning()) {
        return;
      }
      return;
    }

    const pendingId = mainPanelStore.enqueueCommitted(evt.text);

    if (!voiceSession.isRunning()) {
      return;
    }

    if (pendingId === null) {
      return;
    }

    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      const message =
        "Open a text editor so Vocode can run actions against the active file.";
      mainPanelStore.markError(pendingId, message);
      void vscode.window.showWarningMessage(message);
      return;
    }

    const activeFile = editor.document.uri.fsPath;
    const text = evt.text;

    if (inFlightTranscripts === 0) {
      voiceStatus.setProcessing();
    }
    inFlightTranscripts++;

    mainPanelStore.markProcessing(pendingId);

    const pos = editor.selection.active;
    const baseParams = {
      text,
      activeFile,
      workspaceRoot: transcriptWorkspaceRoot(activeFile),
      cursorPosition: { line: pos.line, character: pos.character },
      contextSessionId: voiceSession.contextSessionId(),
    };

    const maxAutoRepairRpcs = Math.max(
      1,
      vscode.workspace
        .getConfiguration("vocode")
        .get<number>("maxTranscriptRepairRpcs", 8),
    );

    void (async () => {
      try {
        const { lastResult, lastOutcomes, reachedLimit } =
          await runRepairChainQueued({
            client,
            baseParams,
            activeFile,
            maxRepairRpcs: maxAutoRepairRpcs,
          });

        const firstFailed = lastOutcomes.find((o) => o.status === "failed");

        if (!lastResult.success) {
          mainPanelStore.markError(pendingId, FAILED_TO_PROCESS_TRANSCRIPT);
          return;
        }

        if (reachedLimit) {
          mainPanelStore.markError(pendingId, "Auto-repair limit reached.");
          return;
        }

        if (firstFailed) {
          const msg =
            firstFailed.message && firstFailed.message !== "not attempted"
              ? firstFailed.message
              : "A directive failed to apply.";
          mainPanelStore.markError(pendingId, msg);
          return;
        }

        mainPanelStore.markHandled(pendingId, {
          summary: lastResult.summary?.trim() || undefined,
        });
      } catch (err) {
        const message =
          err instanceof Error
            ? err.message
            : "Unknown error while running the transcript.";
        mainPanelStore.markError(pendingId, message);
      } finally {
        inFlightTranscripts = Math.max(0, inFlightTranscripts - 1);
        if (voiceSession.isRunning() && inFlightTranscripts === 0) {
          voiceStatus.setListening();
        }
      }
    })();
  });
}
