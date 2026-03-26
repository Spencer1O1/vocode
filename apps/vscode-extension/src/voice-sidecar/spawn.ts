import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import * as fs from "node:fs";
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

  // PortAudio is dynamically linked. When using MSYS2/MinGW, the PortAudio
  // runtime DLLs are typically in `<msys2Root>/mingw64/bin`, which needs to be
  // on PATH for Windows loader resolution.
  const msysRoot = process.env.MSYS2_ROOT ?? "C:\\tools\\msys64";
  const mingw64Bin = path.join(msysRoot, "mingw64", "bin");
  const env = { ...process.env };
  if (process.platform === "win32" && fs.existsSync(mingw64Bin)) {
    env.PATH = `${mingw64Bin};${env.PATH ?? ""}`;
  }

  const proc = spawn(binaryPath, [], {
    cwd: path.dirname(binaryPath),
    stdio: "pipe",
    env,
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
