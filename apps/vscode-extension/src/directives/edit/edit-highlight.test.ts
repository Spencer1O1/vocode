import assert from "node:assert/strict";
import test from "node:test";

import {
  type HighlightLocationInput,
  toLineRanges,
} from "./edit-highlight-ranges";

test("toLineRanges skips locations with missing bounds", () => {
  const locations: HighlightLocationInput[] = [
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 2 },
      selectionEnd: { line: 4 },
    },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 8 },
    },
  ];

  assert.deepEqual(toLineRanges(locations), [
    {
      path: "/tmp/a.ts",
      startLine: 2,
      endLine: 4,
    },
  ]);
});

test("toLineRanges normalizes reversed selections", () => {
  const locations: HighlightLocationInput[] = [
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 10 },
      selectionEnd: { line: 3 },
    },
  ];

  assert.deepEqual(toLineRanges(locations), [
    {
      path: "/tmp/a.ts",
      startLine: 3,
      endLine: 10,
    },
  ]);
});

test("toLineRanges sorts and merges overlapping/adjacent ranges per path", () => {
  const locations: HighlightLocationInput[] = [
    {
      path: "/tmp/b.ts",
      selectionStart: { line: 20 },
      selectionEnd: { line: 25 },
    },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 5 },
      selectionEnd: { line: 8 },
    },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 1 },
      selectionEnd: { line: 3 },
    },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 4 },
      selectionEnd: { line: 4 },
    },
    {
      path: "/tmp/b.ts",
      selectionStart: { line: 24 },
      selectionEnd: { line: 30 },
    },
  ];

  assert.deepEqual(toLineRanges(locations), [
    {
      path: "/tmp/a.ts",
      startLine: 1,
      endLine: 8,
    },
    {
      path: "/tmp/b.ts",
      startLine: 20,
      endLine: 30,
    },
  ]);
});

test("toLineRanges does not merge separated ranges", () => {
  const locations: HighlightLocationInput[] = [
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 1 },
      selectionEnd: { line: 2 },
    },
    {
      path: "/tmp/a.ts",
      selectionStart: { line: 4 },
      selectionEnd: { line: 6 },
    },
  ];

  assert.deepEqual(toLineRanges(locations), [
    {
      path: "/tmp/a.ts",
      startLine: 1,
      endLine: 2,
    },
    {
      path: "/tmp/a.ts",
      startLine: 4,
      endLine: 6,
    },
  ]);
});
