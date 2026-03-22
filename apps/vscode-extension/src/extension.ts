import * as vscode from "vscode";

import { spawnDaemon } from "./daemon/spawn";

function tryStartDaemon(context: vscode.ExtensionContext): boolean {
  try {
    const daemon = spawnDaemon(context);
    console.log(`Vocode daemon started from ${daemon.binaryPath}`);
    return true;
  } catch (err) {
    const message =
      err instanceof Error ? err.message : "Unknown daemon startup error";

    console.error(message);
    void vscode.window.showErrorMessage(
      `Failed to start Vocode daemon: ${message}`,
    );
  }
  return false;
}

export function activate(context: vscode.ExtensionContext) {
  console.log("Vocode extension activated");

  const daemonStarted = tryStartDaemon(context);

  const startVoice = vscode.commands.registerCommand(
    "vocode.startVoice",
    () => {
      vscode.window.showInformationMessage(
        daemonStarted ? "Vocode: Start Voice" : "Daemon not running",
      );
    },
  );

  const stopVoice = vscode.commands.registerCommand("vocode.stopVoice", () => {
    vscode.window.showInformationMessage("Vocode: Stop Voice");
  });

  const applyEdit = vscode.commands.registerCommand("vocode.applyEdit", () => {
    vscode.window.showInformationMessage("Vocode: Apply Edit");
  });

  const runCommand = vscode.commands.registerCommand(
    "vocode.runCommand",
    () => {
      vscode.window.showInformationMessage("Vocode: Run Command");
    },
  );

  context.subscriptions.push(startVoice, stopVoice, applyEdit, runCommand);
}

export function deactivate() {}
