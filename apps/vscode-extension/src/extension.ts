import * as vscode from "vscode";

import { DaemonClient } from "./client/daemon-client";
import { registerAllCommands } from "./commands";
import type { ExtensionServices } from "./commands/services";
import { spawnDaemon } from "./daemon/spawn";

function createServices(context: vscode.ExtensionContext): ExtensionServices {
  try {
    const daemon = spawnDaemon(context);
    console.log(`Vocode daemon started from ${daemon.binaryPath}`);

    return {
      client: new DaemonClient(daemon.process),
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
    };
  }
}

export function activate(context: vscode.ExtensionContext) {
  console.log("Vocode extension activated");

  const services = createServices(context);

  context.subscriptions.push(...registerAllCommands(services), {
    dispose: () => {
      services.client?.dispose();
    },
  });
}

export function deactivate() {}
