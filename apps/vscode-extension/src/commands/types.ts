import type { DaemonClient } from "../client/daemon-client";

export type CommandDefinition =
  | {
      id: string;
      requiresDaemon: false;
      run: () => void | Promise<void>;
    }
  | {
      id: string;
      requiresDaemon: true;
      run: (client: DaemonClient) => void | Promise<void>;
    };
