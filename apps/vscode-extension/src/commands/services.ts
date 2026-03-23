import type { DaemonClient } from "../client/daemon-client";

export interface ExtensionServices {
  client: DaemonClient | null;
}
