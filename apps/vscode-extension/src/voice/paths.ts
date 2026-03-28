import * as fs from "node:fs";
import * as path from "node:path";
import type * as vscode from "vscode";

function getPlatformBinaryName(): string {
  if (process.platform === "win32") return "vocode-voiced.exe";
  return "vocode-voiced";
}

function getPlatformBinarySubdir(): string {
  return `${process.platform}-${process.arch}`;
}

export function getDevVoiceSidecarPath(
  context: vscode.ExtensionContext,
): string {
  return path.resolve(
    context.extensionPath,
    "..",
    "..",
    "apps",
    "voice",
    "bin",
    getPlatformBinarySubdir(),
    getPlatformBinaryName(),
  );
}

export function getProdVoiceSidecarPath(
  context: vscode.ExtensionContext,
): string {
  return path.join(
    context.extensionPath,
    "bin",
    getPlatformBinarySubdir(),
    getPlatformBinaryName(),
  );
}

export function resolveVoiceSidecarPath(
  context: vscode.ExtensionContext,
): string {
  const devPath = getDevVoiceSidecarPath(context);
  if (fs.existsSync(devPath)) {
    console.log(`[vocode] using dev voice sidecar: ${devPath}`);
    return devPath;
  }

  const prodPath = getProdVoiceSidecarPath(context);
  if (fs.existsSync(prodPath)) {
    console.log(`[vocode] using bundled voice sidecar: ${prodPath}`);
    return prodPath;
  }

  throw new Error(
    `Could not locate Vocode voice sidecar binary for ${getPlatformBinarySubdir()}`,
  );
}
