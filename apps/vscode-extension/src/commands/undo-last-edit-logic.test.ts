import assert from "node:assert/strict";
import test from "node:test";

import { runUndoLastEditWithDeps } from "./undo-last-edit-logic";

test("runUndoLastEditWithDeps runs undo when editor is active", async () => {
  let executedUndo = false;
  let warning = "";

  await runUndoLastEditWithDeps({
    hasActiveEditor: () => true,
    executeUndo: () => {
      executedUndo = true;
      return Promise.resolve();
    },
    showWarning: (message) => {
      warning = message;
    },
  });

  assert.equal(executedUndo, true);
  assert.equal(warning, "");
});

test("runUndoLastEditWithDeps warns and skips undo without active editor", async () => {
  let executedUndo = false;
  let warning = "";

  await runUndoLastEditWithDeps({
    hasActiveEditor: () => false,
    executeUndo: () => {
      executedUndo = true;
      return Promise.resolve();
    },
    showWarning: (message) => {
      warning = message;
    },
  });

  assert.equal(executedUndo, false);
  assert.match(warning, /Open a text editor/);
});
