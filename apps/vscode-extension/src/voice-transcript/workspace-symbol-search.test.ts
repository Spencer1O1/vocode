import assert from "node:assert/strict";
import test from "node:test";

import {
  buildWorkspaceSymbolQueryVariants,
  expandIdentifierTokens,
  symbolKindHintMatches,
  symbolMatchesNaturalLanguageQuery,
} from "./workspace-symbol-search";

// vscode.SymbolKind at runtime in tests (no vscode module in node)
const SymbolKindFunction = 11;
const SymbolKindVariable = 12;

test("expandIdentifierTokens splits camelCase and snake_case", () => {
  assert.equal(expandIdentifierTokens("deltaTime"), "delta time");
  assert.equal(expandIdentifierTokens("foo_bar"), "foo bar");
});

test("buildWorkspaceSymbolQueryVariants prefers camelCase for two words", () => {
  const v = buildWorkspaceSymbolQueryVariants("delta time");
  assert.equal(v[0], "deltaTime");
  assert.ok(v.includes("deltatime"));
  assert.ok(v.includes("delta time"));
  assert.ok(v.includes("delta"));
});

test("symbolMatchesNaturalLanguageQuery matches phrase to camelCase symbol", () => {
  assert.ok(symbolMatchesNaturalLanguageQuery("delta time", "deltaTime", ""));
  assert.ok(symbolMatchesNaturalLanguageQuery("Delta Time", "deltaTime", ""));
});

test("symbolMatchesNaturalLanguageQuery is case tolerant on symbol", () => {
  assert.ok(symbolMatchesNaturalLanguageQuery("thing", "Thing", ""));
  assert.ok(symbolMatchesNaturalLanguageQuery("Thing", "thing", ""));
});

test("symbolMatchesNaturalLanguageQuery uses container", () => {
  assert.ok(
    symbolMatchesNaturalLanguageQuery(
      "delta time",
      "x",
      "Helper with deltaTime inside",
    ),
  );
});

test("symbolKindHintMatches filters by classifier hint", () => {
  assert.ok(symbolKindHintMatches("", SymbolKindFunction));
  assert.ok(symbolKindHintMatches("function", SymbolKindFunction));
  assert.equal(symbolKindHintMatches("function", SymbolKindVariable), false);
  assert.ok(symbolKindHintMatches("bogus_hint", SymbolKindVariable));
});
