import assert from "node:assert/strict";
import test from "node:test";

import type { AppliedEditLocation } from "./dispatch-workspace-edit";
import { toLineRanges } from "./edit-highlight-ranges";

function loc(
  path: string,
  startLine: number,
  endLine = startLine,
): AppliedEditLocation {
  return {
    path,
    selectionStart: { line: startLine } as { line: number } as never,
    selectionEnd: { line: endLine } as { line: number } as never,
  };
}

test("toLineRanges drops edits without start/end positions", () => {
  const ranges = toLineRanges([
    { path: "/tmp/a.ts" },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 4 } as { line: number } as never,
    },
  ]);

  assert.deepEqual(ranges, []);
});

test("toLineRanges normalizes reversed line selections", () => {
  const ranges = toLineRanges([loc("/tmp/a.ts", 9, 4)]);

  assert.deepEqual(ranges, [{ path: "/tmp/a.ts", startLine: 4, endLine: 9 }]);
});

test("toLineRanges merges overlapping and adjacent ranges per file", () => {
  const ranges = toLineRanges([
    loc("/tmp/a.ts", 2, 4),
    loc("/tmp/a.ts", 5, 8),
    loc("/tmp/a.ts", 7, 10),
    loc("/tmp/a.ts", 15, 15),
  ]);

  assert.deepEqual(ranges, [
    { path: "/tmp/a.ts", startLine: 2, endLine: 10 },
    { path: "/tmp/a.ts", startLine: 15, endLine: 15 },
  ]);
});

test("toLineRanges does not merge ranges across different files", () => {
  const ranges = toLineRanges([
    loc("/tmp/b.ts", 1, 2),
    loc("/tmp/a.ts", 3, 4),
    loc("/tmp/b.ts", 3, 5),
  ]);

  assert.deepEqual(ranges, [
    { path: "/tmp/a.ts", startLine: 3, endLine: 4 },
    { path: "/tmp/b.ts", startLine: 1, endLine: 5 },
  ]);
});
