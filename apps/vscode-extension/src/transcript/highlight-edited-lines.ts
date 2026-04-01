import * as vscode from "vscode";

import type { EditLocationMap } from "../directives/navigation/execute-navigation-intent";

const HIGHLIGHT_DURATION_MS = 3500;

const editedLineDecoration = vscode.window.createTextEditorDecorationType({
  isWholeLine: true,
  backgroundColor: "rgba(56, 189, 248, 0.18)",
  borderRadius: "2px",
  overviewRulerColor: "rgba(56, 189, 248, 0.85)",
  overviewRulerLane: vscode.OverviewRulerLane.Right,
});

let clearHighlightTimer: ReturnType<typeof setTimeout> | undefined;

function lineRange(line: number): vscode.Range {
  const start = new vscode.Position(line, 0);
  return new vscode.Range(start, start);
}

function collectEditedLines(
  editLocations: EditLocationMap,
): Map<string, Set<number>> {
  const editedLinesByPath = new Map<string, Set<number>>();

  for (const editId in editLocations) {
    const location = editLocations[editId];
    if (!location) {
      continue;
    }
    const line = location.selectionStart?.line;
    if (line === undefined) {
      continue;
    }
    const lines = editedLinesByPath.get(location.path) ?? new Set<number>();
    lines.add(line);
    editedLinesByPath.set(location.path, lines);
  }

  return editedLinesByPath;
}

/**
 * Highlights transcript-applied edited lines in visible editors for a short duration.
 */
export function highlightEditedLines(editLocations: EditLocationMap): void {
  const editedLinesByPath = collectEditedLines(editLocations);
  if (editedLinesByPath.size === 0) {
    return;
  }

  for (const editor of vscode.window.visibleTextEditors) {
    const lines = editedLinesByPath.get(editor.document.uri.fsPath);
    if (!lines || lines.size === 0) {
      continue;
    }
    const ranges = [...lines].sort((a, b) => a - b).map(lineRange);
    editor.setDecorations(editedLineDecoration, ranges);
  }

  if (clearHighlightTimer) {
    clearTimeout(clearHighlightTimer);
  }
  clearHighlightTimer = setTimeout(() => {
    for (const editor of vscode.window.visibleTextEditors) {
      editor.setDecorations(editedLineDecoration, []);
    }
    clearHighlightTimer = undefined;
  }, HIGHLIGHT_DURATION_MS);
}
