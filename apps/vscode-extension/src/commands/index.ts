import type * as vscode from "vscode";

import { applyEditCommand } from "./apply-edit";
import { registerCommands } from "./helpers";
import { pingCommand } from "./ping";
import { runCommand } from "./run-command";
import type { ExtensionServices } from "./services";
import { startVoiceCommand } from "./start-voice";
import { stopVoiceCommand } from "./stop-voice";
import type { CommandDefinition } from "./types";

const definitions: CommandDefinition[] = [
  pingCommand,
  applyEditCommand,
  startVoiceCommand,
  stopVoiceCommand,
  runCommand,
];

export function registerAllCommands(
  services: ExtensionServices,
): vscode.Disposable[] {
  return registerCommands(services, definitions);
}
