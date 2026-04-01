export interface LinePositionLike {
  line: number;
}

export interface HighlightLocationInput {
  path: string;
  selectionStart?: LinePositionLike;
  selectionEnd?: LinePositionLike;
}

export interface HighlightLineRange {
  path: string;
  startLine: number;
  endLine: number;
}

/**
 * Normalizes, sorts, and merges line ranges for highlight decorations.
 * Adjacent ranges are merged so users see one contiguous block per file.
 */
export function toLineRanges(
  locations: readonly HighlightLocationInput[],
): HighlightLineRange[] {
  const normalized: HighlightLineRange[] = [];

  for (const loc of locations) {
    const startLine = loc.selectionStart?.line;
    const endLine = loc.selectionEnd?.line;
    if (startLine === undefined || endLine === undefined) {
      continue;
    }

    normalized.push({
      path: loc.path,
      startLine: Math.min(startLine, endLine),
      endLine: Math.max(startLine, endLine),
    });
  }

  normalized.sort((a, b) => {
    if (a.path === b.path) {
      if (a.startLine === b.startLine) {
        return a.endLine - b.endLine;
      }
      return a.startLine - b.startLine;
    }
    return a.path.localeCompare(b.path);
  });

  const merged: HighlightLineRange[] = [];
  for (const range of normalized) {
    const prev = merged.at(-1);
    if (
      !prev ||
      prev.path !== range.path ||
      range.startLine > prev.endLine + 1
    ) {
      merged.push({ ...range });
      continue;
    }

    prev.endLine = Math.max(prev.endLine, range.endLine);
  }

  return merged;
}
