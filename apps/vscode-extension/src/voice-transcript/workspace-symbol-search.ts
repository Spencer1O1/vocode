/**
 * Helpers for host workspace symbol search: voice queries are often multi-word
 * ("delta time") while VS Code / LSP returns best results for identifier-shaped
 * queries; we derive variants and match case-insensitively with camelCase awareness.
 */

/** vscode.SymbolKind numeric values (extension host only; kept numeric so node tests need no vscode). */
const SK = {
  Module: 1,
  Namespace: 2,
  Package: 3,
  Class: 4,
  Method: 5,
  Property: 6,
  Field: 7,
  Constructor: 8,
  Enum: 9,
  Interface: 10,
  Function: 11,
  Variable: 12,
  Constant: 13,
  Struct: 22,
  TypeParameter: 25,
} as const;

/** Insert word boundaries for camelCase / PascalCase / snake_case token matching. */
export function expandIdentifierTokens(s: string): string {
  const withBreaks = s
    .replace(/([a-z\d])([A-Z])/g, "$1 $2")
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1 $2");
  return withBreaks
    .replace(/_/g, " ")
    .toLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}

function compactAlnum(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]/g, "");
}

/**
 * True if the symbol's name/container matches the user's natural-language query.
 * - Multi-word: each word must appear in an expanded token view (e.g. deltaTime → "delta time").
 * - Single run-together token: also allow compact substring match (e.g. "deltatime" vs deltaTime).
 */
export function symbolMatchesNaturalLanguageQuery(
  query: string,
  name: string,
  containerName: string,
): boolean {
  const q = query.trim().toLowerCase();
  if (!q) {
    return false;
  }
  const expanded =
    `${expandIdentifierTokens(name)} ${expandIdentifierTokens(containerName)}`.replace(
      /\s+/g,
      " ",
    );
  const compact = compactAlnum(name + containerName);
  const parts = q.split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return false;
  }
  return parts.every((p) => {
    const pl = p.toLowerCase();
    if (expanded.includes(pl)) {
      return true;
    }
    const pCompact = compactAlnum(pl);
    return pCompact.length >= 2 && compact.includes(pCompact);
  });
}

/**
 * Query strings to pass to executeWorkspaceSymbolProvider, most specific first.
 * Multi-word phrases become camelCase / compact forms so the LSP returns candidates.
 */
export function buildWorkspaceSymbolQueryVariants(query: string): string[] {
  const t = query.trim();
  if (!t) {
    return [];
  }
  const words = t.split(/\s+/).filter(Boolean);
  const variants: string[] = [];

  if (words.length >= 2) {
    const camel =
      words[0].toLowerCase() +
      words
        .slice(1)
        .map(
          (w) =>
            w.charAt(0).toUpperCase() +
            (w.length > 1 ? w.slice(1).toLowerCase() : ""),
        )
        .join("");
    variants.push(camel);
  }

  variants.push(words.join(""));
  variants.push(t);

  if (words.length > 0 && words[0]) {
    variants.push(words[0]);
  }

  const seen = new Set<string>();
  const out: string[] = [];
  for (const v of variants) {
    const k = v.trim();
    if (!k || seen.has(k)) {
      continue;
    }
    seen.add(k);
    out.push(k);
  }
  return out;
}

const KIND_BY_HINT: Record<string, readonly number[]> = {
  function: [SK.Function],
  method: [SK.Method],
  class: [SK.Class],
  variable: [SK.Variable],
  constant: [SK.Constant, SK.Variable],
  interface: [SK.Interface],
  enum: [SK.Enum],
  property: [SK.Property],
  field: [SK.Field],
  constructor: [SK.Constructor],
  module: [SK.Module, SK.Namespace, SK.Package],
  struct: [SK.Struct],
  type: [SK.Interface, SK.Class, SK.TypeParameter, SK.Struct],
};

/** When hint is empty/any/unknown, or unrecognized, do not filter by LSP kind. */
export function symbolKindHintMatches(
  hint: string | undefined,
  kind: number,
): boolean {
  const h = hint?.trim().toLowerCase() ?? "";
  if (h === "" || h === "any" || h === "unknown") {
    return true;
  }
  const allowed = KIND_BY_HINT[h];
  if (!allowed?.length) {
    return true;
  }
  return allowed.includes(kind);
}
