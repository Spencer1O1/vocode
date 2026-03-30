import type { AppliedEditLocation } from "./dispatch-workspace-edit";

export interface HighlightLineRange {
  path: string;
  startLine: number;
  endLine: number;
}

export function toLineRanges(
  locations: AppliedEditLocation[],
): HighlightLineRange[] {
  const ranges: HighlightLineRange[] = [];
  for (const loc of locations) {
    const startLine = loc.selectionStart?.line;
    const endLine = loc.selectionEnd?.line;
    if (startLine === undefined || endLine === undefined) {
      continue;
    }
    ranges.push({
      path: loc.path,
      startLine: Math.min(startLine, endLine),
      endLine: Math.max(startLine, endLine),
    });
  }

  ranges.sort((a, b) => {
    const pathCmp = a.path.localeCompare(b.path);
    if (pathCmp !== 0) return pathCmp;
    if (a.startLine !== b.startLine) return a.startLine - b.startLine;
    return a.endLine - b.endLine;
  });

  const merged: HighlightLineRange[] = [];
  for (const range of ranges) {
    const prev = merged[merged.length - 1];
    if (
      prev &&
      prev.path === range.path &&
      range.startLine <= prev.endLine + 1
    ) {
      prev.endLine = Math.max(prev.endLine, range.endLine);
      continue;
    }
    merged.push({ ...range });
  }

  return merged;
}
