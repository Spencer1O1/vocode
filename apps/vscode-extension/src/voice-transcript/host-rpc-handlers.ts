import type {
  HostGetDocumentSymbolsParams,
  HostGetDocumentSymbolsResult,
  HostReadFileParams,
  HostReadFileResult,
  HostWorkspaceSymbolSearchParams,
  HostWorkspaceSymbolSearchResult,
} from "@vocode/protocol";
import * as vscode from "vscode";

import { flattenDocumentSymbols } from "./document-symbols";
import {
  buildWorkspaceSymbolQueryVariants,
  symbolKindHintMatches,
  symbolMatchesNaturalLanguageQuery,
} from "./workspace-symbol-search";

export async function handleHostReadFile(
  params: HostReadFileParams,
): Promise<HostReadFileResult> {
  const uri = vscode.Uri.file(params.path);
  const bytes = await vscode.workspace.fs.readFile(uri);
  return { text: new TextDecoder("utf-8").decode(bytes) };
}

export async function handleHostGetDocumentSymbols(
  params: HostGetDocumentSymbolsParams,
): Promise<HostGetDocumentSymbolsResult> {
  const uri = vscode.Uri.file(params.path);
  const raw = await vscode.commands.executeCommand<
    vscode.DocumentSymbol[] | undefined
  >("vscode.executeDocumentSymbolProvider", uri);
  return { symbols: flattenDocumentSymbols(raw) };
}

function symbolKey(s: vscode.SymbolInformation): string {
  const start = s.location.range.start;
  return `${s.location.uri.fsPath}:${start.line}:${start.character}:${s.name ?? ""}`;
}

export async function handleHostWorkspaceSymbolSearch(
  params: HostWorkspaceSymbolSearchParams,
): Promise<HostWorkspaceSymbolSearchResult> {
  const variants = buildWorkspaceSymbolQueryVariants(params.query);
  const seenKeys = new Set<string>();
  const candidates: vscode.SymbolInformation[] = [];
  const maxHits = 20;
  const maxRawPerVariant = 400;

  for (const variant of variants) {
    const raw = await vscode.commands.executeCommand<
      vscode.SymbolInformation[] | undefined
    >("vscode.executeWorkspaceSymbolProvider", variant);
    let rawSeen = 0;
    for (const s of raw ?? []) {
      rawSeen++;
      if (rawSeen > maxRawPerVariant) {
        break;
      }
      const name = s.name ?? "";
      const container = s.containerName ?? "";
      if (!symbolMatchesNaturalLanguageQuery(params.query, name, container)) {
        continue;
      }
      if (!symbolKindHintMatches(params.symbolKind, s.kind)) {
        continue;
      }
      const k = symbolKey(s);
      if (seenKeys.has(k)) {
        continue;
      }
      seenKeys.add(k);
      candidates.push(s);
      if (candidates.length >= maxHits) {
        break;
      }
    }
    if (candidates.length >= maxHits) {
      break;
    }
  }

  const hits: HostWorkspaceSymbolSearchResult["hits"] = [];
  for (const s of candidates) {
    const name = s.name ?? "";
    const path = s.location.uri.fsPath;
    const r = s.location.range;
    const start = r.start;
    const end = r.end;
    const matchLength =
      start.line === end.line
        ? Math.max(1, end.character - start.character)
        : 1;
    hits.push({
      path,
      line: start.line,
      character: start.character,
      preview: name,
      matchLength,
    });
  }
  return { hits };
}
