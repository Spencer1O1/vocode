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
    const filePath = editor.document.uri.fsPath;
    const fileRanges = groupedRanges.filter((range) => range.path === filePath);
    if (fileRanges.length === 0) {
      continue;
    }

    editor.setDecorations(
      decoration,
      fileRanges.map((range) => {
        const line = editor.document.lineAt(range.endLine);
        return new vscode.Range(
          range.startLine,
          0,
          range.endLine,
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
