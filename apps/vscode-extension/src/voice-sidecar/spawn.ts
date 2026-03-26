import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import * as path from "node:path";
import type * as vscode from "vscode";

import { resolveVoiceSidecarPath } from "./paths";

export interface SpawnedVoiceSidecar {
  process: ChildProcessWithoutNullStreams;
  binaryPath: string;
}

export function spawnVoiceSidecar(
  context: vscode.ExtensionContext,
): SpawnedVoiceSidecar {
  const binaryPath = resolveVoiceSidecarPath(context);

  const proc = spawn(binaryPath, [], {
    cwd: path.dirname(binaryPath),
    stdio: "pipe",
  });

  proc.stdout.on("data", (data: Buffer) => {
    console.log("[vocode-voiced stdout]", data.toString());
  });

  proc.stderr.on("data", (data: Buffer) => {
    console.error(`[vocode-voiced stderr] ${data.toString()}`);
  });

  proc.on("error", (error: Error) => {
    console.error(`[vocode-voiced spawn error] ${error.message}`);
  });

  proc.on("exit", (code: number | null, signal: NodeJS.Signals | null) => {
    console.log(`vocode-voiced exited with code=${code} signal=${signal}`);
  });

  context.subscriptions.push({
    dispose: () => {
      if (!proc.killed) {
        proc.kill();
      }
    },
  });

  return {
    process: proc,
    binaryPath,
  };
}
