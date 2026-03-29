import * as vscode from "vscode";

/** Max keywords merged from all workspace `.vocode` files (before JSON env to sidecar). */
const MAX_WORKSPACE_STT_KEYWORDS = 80;

export type WorkspaceVocodeShape = {
  sttKeywords?: unknown;
};

/**
 * Reads `.vocode` at each workspace folder root (JSON), merges `sttKeywords` string entries in
 * folder order, de-duplicated case-insensitively.
 */
export async function readWorkspaceSttKeywords(): Promise<string[]> {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders?.length) {
    return [];
  }
  const merged: string[] = [];
  const seen = new Set<string>();
  for (const folder of folders) {
    const uri = vscode.Uri.joinPath(folder.uri, ".vocode");
    try {
      const data = await vscode.workspace.fs.readFile(uri);
      const text = new TextDecoder("utf-8").decode(data);
      let parsed: unknown;
      try {
        parsed = JSON.parse(text) as unknown;
      } catch {
        console.warn(
          `[vocode] Invalid JSON in ${uri.fsPath}; ignoring workspace STT keywords.`,
        );
        continue;
      }
      if (!parsed || typeof parsed !== "object") {
        continue;
      }
      const k = (parsed as WorkspaceVocodeShape).sttKeywords;
      if (!Array.isArray(k)) {
        continue;
      }
      for (const item of k) {
        if (typeof item !== "string") {
          continue;
        }
        const t = item.trim();
        if (!t) {
          continue;
        }
        const lk = t.toLowerCase();
        if (seen.has(lk)) {
          continue;
        }
        seen.add(lk);
        merged.push(t);
        if (merged.length >= MAX_WORKSPACE_STT_KEYWORDS) {
          return merged;
        }
      }
    } catch {
      // Missing file or unreadable — skip this folder.
    }
  }
  return merged;
}

/**
 * Creates `.vocode` in the first workspace folder with `{"sttKeywords":[]}` when none exists.
 */
export async function createWorkspaceVocodeFile(): Promise<void> {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders?.length) {
    void vscode.window.showErrorMessage(
      "Open a folder or multi-root workspace first to create .vocode.",
    );
    return;
  }
  const uri = vscode.Uri.joinPath(folders[0].uri, ".vocode");
  try {
    await vscode.workspace.fs.stat(uri);
    void vscode.window.showWarningMessage(
      `.vocode already exists: ${uri.fsPath}`,
    );
    await vscode.window.showTextDocument(uri);
    return;
  } catch {
    // create
  }
  const body = `${JSON.stringify({ sttKeywords: [] }, null, 2)}\n`;
  await vscode.workspace.fs.writeFile(uri, new TextEncoder().encode(body));
  void vscode.window.showInformationMessage(
    "Created .vocode with sttKeywords. Edit the file, then use “Apply changes and restart”.",
  );
  await vscode.window.showTextDocument(uri);
}
