import * as vscode from "vscode";

import { getVocodeSetupBlockReason } from "../config/spawn-env";
import type { MainPanelViewProvider } from "../ui/panel/main-panel";
import type { ExtensionServices } from "./services";
import type { CommandDefinition } from "./types";

export type RegisterCommandsOptions = {
  extensionContext: vscode.ExtensionContext;
  mainPanel: MainPanelViewProvider;
};

export function registerCommands(
  services: ExtensionServices,
  definitions: CommandDefinition[],
  options?: RegisterCommandsOptions,
): vscode.Disposable[] {
  return definitions.map((definition) =>
    vscode.commands.registerCommand(definition.id, async () => {
      try {
        if (definition.requiresDaemon) {
          if (!services.client) {
            if (options !== undefined) {
              const block = await getVocodeSetupBlockReason(
                options.extensionContext,
              );
              if (block !== null) {
                options.mainPanel.revealPanelView("settings");
                void vscode.window
                  .showErrorMessage(block, "Focus Vocode panel")
                  .then((choice) => {
                    if (choice === "Focus Vocode panel") {
                      options.mainPanel.revealPanelView("settings");
                    }
                  });
                return;
              }
            }
            void vscode.window.showErrorMessage("Vocode core is not running.");
            return;
          }

          await definition.run(services.client, services);
          return;
        }

        await definition.run(services);
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Unknown command error";

        console.error(`[vocode] command ${definition.id} failed:`, error);
        void vscode.window.showErrorMessage(
          `${definition.id} failed: ${message}`,
        );
      }
    }),
  );
}
