import * as fs from "node:fs";
import * as path from "node:path";
import type * as vscode from "vscode";

export function getPlatformBinaryName(): string {
  if (process.platform === "win32") return "vocoded.exe";
  return "vocoded";
}

export function getPlatformBinarySubdir(): string {
  return `${process.platform}-${process.arch}`;
}

export function getDevDaemonPath(context: vscode.ExtensionContext): string {
  return path.resolve(
    context.extensionPath,
    "..",
    "..",
    "apps",
    "daemon",
    "bin",
    getPlatformBinarySubdir(),
    getPlatformBinaryName(),
  );
}

export function getProdDaemonPath(context: vscode.ExtensionContext): string {
  return path.join(
    context.extensionPath,
    "bin",
    getPlatformBinarySubdir(),
    getPlatformBinaryName(),
  );
}

export function resolveDaemonPath(context: vscode.ExtensionContext): string {
  const devPath = getDevDaemonPath(context);
  if (fs.existsSync(devPath)) {
    console.log(`[vocode] using dev daemon: ${devPath}`);
    return devPath;
  }

  const prodPath = getProdDaemonPath(context);
  if (fs.existsSync(prodPath)) {
    console.log(`[vocode] using bundled daemon: ${prodPath}`);
    return prodPath;
  }

  throw new Error(
    `Could not locate Vocode daemon binary for ${getPlatformBinarySubdir()}`,
  );
}
