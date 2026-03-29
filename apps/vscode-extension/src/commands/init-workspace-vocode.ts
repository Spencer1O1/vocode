import { createWorkspaceVocodeFile } from "../config/workspace-vocode";
import type { CommandDefinition } from "./types";

export const initWorkspaceVocodeCommand: CommandDefinition = {
  id: "vocode.initWorkspaceVocode",
  requiresDaemon: false,
  run: async () => {
    await createWorkspaceVocodeFile();
  },
};
