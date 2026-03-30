import * as vscode from "vscode";

import type { AppliedEditLocation } from "./dispatch-workspace-edit";
import { toLineRanges } from "./edit-highlight-ranges";

const HIGHLIGHT_DURATION_MS = 2500;

let decorationType: vscode.TextEditorDecorationType | undefined;
let clearTimer: NodeJS.Timeout | undefined;

function getDecorationType(): vscode.TextEditorDecorationType {
  if (!decorationType) {
    decorationType = vscode.window.createTextEditorDecorationType({
      isWholeLine: true,
      backgroundColor: new vscode.ThemeColor("editor.rangeHighlightBackground"),
      overviewRulerColor: new vscode.ThemeColor(
        "editor.findMatchHighlightBackground",
      ),
      overviewRulerLane: vscode.OverviewRulerLane.Right,
    });
  }
  return decorationType;
}

/** Temporarily highlights edited lines in currently visible editors. */
export function highlightEditedLines(locations: AppliedEditLocation[]): void {
  const groupedRanges = toLineRanges(locations);
  if (groupedRanges.length === 0) {
    return;
  }

  const decoration = getDecorationType();
  for (const editor of vscode.window.visibleTextEditors) {
    const document = editor.document;
    const filePath = document.uri.fsPath;
    const fileRanges = groupedRanges.filter((range) => range.path === filePath);
    if (fileRanges.length === 0) {
      continue;
    }

    const lastLineIndex = Math.max(0, document.lineCount - 1);

    editor.setDecorations(
      decoration,
      fileRanges.map((range) => {
        // Clamp start/end lines to the current document to avoid stale indices.
        let startLine = Math.min(
          Math.max(range.startLine, 0),
          lastLineIndex,
        );
        let endLine = Math.min(
          Math.max(range.endLine, 0),
          lastLineIndex,
        );

        if (endLine < startLine) {
          startLine = endLine;
        }

        const line = document.lineAt(endLine);
        return new vscode.Range(
          startLine,
          0,
          endLine,
          line.range.end.character,
        );
      }),
    );
  }

  if (clearTimer) {
    clearTimeout(clearTimer);
  }
  clearTimer = setTimeout(() => {
    for (const editor of vscode.window.visibleTextEditors) {
      editor.setDecorations(decoration, []);
    }
  }, HIGHLIGHT_DURATION_MS);
}
