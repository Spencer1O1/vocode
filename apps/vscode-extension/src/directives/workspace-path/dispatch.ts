import * as fs from "node:fs/promises";
import * as path from "node:path";
import type { VoiceTranscriptDirective } from "@vocode/protocol";
import * as vscode from "vscode";

import type { DirectiveDispatchOutcome } from "../dispatch";

function workspaceRootPath(): string | undefined {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}

/** True if target is the workspace root or a path inside it (after resolve). */
function isUnderWorkspaceRoot(root: string, targetPath: string): boolean {
  const rootResolved = path.resolve(root);
  const targetResolved = path.resolve(targetPath);
  if (targetResolved === rootResolved) {
    return true;
  }
  const prefix = rootResolved.endsWith(path.sep)
    ? rootResolved
    : `${rootResolved}${path.sep}`;
  return targetResolved.startsWith(prefix);
}

function isResolvedWorkspaceRoot(root: string, targetPath: string): boolean {
  return path.resolve(targetPath) === path.resolve(root);
}

function fsErrorMessage(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

function resolvedPathsEqual(a: string, b: string): boolean {
  return path.resolve(a) === path.resolve(b);
}

/** Close editor tabs whose document path matches (after a move/rename, the old path is stale). */
async function closeTextEditorTabsForPath(filePath: string): Promise<void> {
  const want = path.resolve(filePath);
  for (const group of vscode.window.tabGroups.all) {
    for (const tab of group.tabs) {
      const input = tab.input;
      if (input instanceof vscode.TabInputText) {
        if (resolvedPathsEqual(input.uri.fsPath, want)) {
          await vscode.window.tabGroups.close(tab);
        }
      }
    }
  }
}

/** Focus the destination path when it is a text file (folders skipped). */
async function revealDestinationIfTextFile(toPath: string): Promise<void> {
  try {
    const st = await fs.stat(toPath);
    if (!st.isFile()) {
      return;
    }
    const doc = await vscode.workspace.openTextDocument(
      vscode.Uri.file(toPath),
    );
    await vscode.window.showTextDocument(doc, { preview: false });
  } catch {
    // Non-text or unreadable — navigation directive may still run next.
  }
}

async function deleteFileUnderWorkspace(
  root: string,
  filePath: string,
): Promise<DirectiveDispatchOutcome> {
  if (!filePath || !isUnderWorkspaceRoot(root, filePath)) {
    return {
      ok: false,
      message: "delete_file: path missing or outside workspace.",
    };
  }
  if (isResolvedWorkspaceRoot(root, filePath)) {
    return {
      ok: false,
      message: "delete_file: cannot delete the workspace folder.",
    };
  }
  try {
    await fs.unlink(filePath);
  } catch (e) {
    return { ok: false, message: `delete_file failed: ${fsErrorMessage(e)}` };
  }
  return { ok: true };
}

async function movePathUnderWorkspace(
  root: string,
  from: string,
  to: string,
): Promise<DirectiveDispatchOutcome> {
  if (
    !from ||
    !to ||
    !isUnderWorkspaceRoot(root, from) ||
    !isUnderWorkspaceRoot(root, to)
  ) {
    return {
      ok: false,
      message: "move_path: paths missing or outside workspace.",
    };
  }
  if (isResolvedWorkspaceRoot(root, from)) {
    return {
      ok: false,
      message: "move_path: cannot move the workspace folder.",
    };
  }
  try {
    await fs.mkdir(path.dirname(to), { recursive: true });
    await fs.rename(from, to);
    await closeTextEditorTabsForPath(from);
    await revealDestinationIfTextFile(to);
  } catch (e) {
    return { ok: false, message: `move_path failed: ${fsErrorMessage(e)}` };
  }
  return { ok: true };
}

async function createFolderUnderWorkspace(
  root: string,
  dirPath: string,
): Promise<DirectiveDispatchOutcome> {
  if (!dirPath || !isUnderWorkspaceRoot(root, dirPath)) {
    return {
      ok: false,
      message: "create_folder: path missing or outside workspace.",
    };
  }
  try {
    await fs.mkdir(dirPath, { recursive: true });
  } catch (e) {
    return {
      ok: false,
      message: `create_folder failed: ${fsErrorMessage(e)}`,
    };
  }
  return { ok: true };
}

/**
 * Applies delete_file / move_path / create_folder under the first workspace folder (host-side jail).
 */
export async function dispatchWorkspacePath(
  d: VoiceTranscriptDirective,
): Promise<DirectiveDispatchOutcome> {
  const root = workspaceRootPath();
  if (!root) {
    return { ok: false, message: "No workspace folder is open." };
  }

  switch (d.kind) {
    case "delete_file":
      return await deleteFileUnderWorkspace(
        root,
        d.deleteFileDirective?.path?.trim() ?? "",
      );
    case "move_path":
      return await movePathUnderWorkspace(
        root,
        d.movePathDirective?.from?.trim() ?? "",
        d.movePathDirective?.to?.trim() ?? "",
      );
    case "create_folder":
      return await createFolderUnderWorkspace(
        root,
        d.createFolderDirective?.path?.trim() ?? "",
      );
    default:
      return {
        ok: false,
        message: "internal: not a workspace path directive",
      };
  }
}
