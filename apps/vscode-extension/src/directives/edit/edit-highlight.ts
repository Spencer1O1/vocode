import path from "node:path";
import * as vscode from "vscode";

import type { HighlightLocationInput } from "./edit-highlight-ranges";
import { toLineRanges } from "./edit-highlight-ranges";

const EDIT_HIGHLIGHT_CLEAR_MS = 2500;

const editedLineDecoration = vscode.window.createTextEditorDecorationType({
  isWholeLine: true,
  backgroundColor: new vscode.ThemeColor("editor.rangeHighlightBackground"),
  overviewRulerColor: new vscode.ThemeColor("editor.rangeHighlightBackground"),
  overviewRulerLane: vscode.OverviewRulerLane.Full,
});

/**
 * Applies temporary whole-line highlights for recently edited lines.
 */
export function highlightEditedLines(
  locations: readonly HighlightLocationInput[],
  clearAfterMs = EDIT_HIGHLIGHT_CLEAR_MS,
): void {
  const ranges = toLineRanges(locations);
  if (ranges.length === 0) {
    return;
  }

  const highlightedEditors: vscode.TextEditor[] = [];

  for (const editor of vscode.window.visibleTextEditors) {
    const editorPath = path.resolve(editor.document.uri.fsPath);
    const editorRanges = ranges
      .filter((range) => path.resolve(range.path) === editorPath)
      .map(
        (range) =>
          new vscode.Range(
            new vscode.Position(range.startLine, 0),
            new vscode.Position(range.endLine, Number.MAX_SAFE_INTEGER),

          ),
      );

    if (editorRanges.length === 0) {
      continue;
    }

    editor.setDecorations(editedLineDecoration, editorRanges);
    highlightedEditors.push(editor);
  }

  if (highlightedEditors.length === 0) {
    return;
  }

  setTimeout(() => {
    for (const editor of highlightedEditors) {
      editor.setDecorations(editedLineDecoration, []);
    }
  }, clearAfterMs);
}
