import assert from "node:assert/strict";
import test from "node:test";

import { isEditDirective } from "./validators";

test("isEditDirective accepts explicit success shape", () => {
  const value = {
    kind: "success",
    actions: [
      {
        kind: "replace_between_anchors",
        path: "/tmp/file.ts",
        anchor: { before: "A", after: "B" },
        newText: "updated",
      },
    ],
  };

  assert.equal(isEditDirective(value), true);
});

test("isEditDirective rejects mixed success/failure shape", () => {
  const value = {
    kind: "success",
    actions: [],
    failure: { code: "validation_failed", message: "bad" },
  };

  assert.equal(isEditDirective(value), false);
});

test("isEditDirective accepts explicit noop shape", () => {
  const value = {
    kind: "noop",
    reason: "No change needed.",
  };

  assert.equal(isEditDirective(value), true);
});

test("isEditDirective rejects extra keys on success", () => {
  const value = {
    kind: "success",
    actions: [],
    extra: true,
  };

  assert.equal(isEditDirective(value), false);
});

test("isEditDirective rejects invalid failure code", () => {
  const value = {
    kind: "failure",
    failure: { code: "not_real", message: "bad" },
  };

  assert.equal(isEditDirective(value), false);
});
