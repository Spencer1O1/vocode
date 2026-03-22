import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import $RefParser from "@apidevtools/json-schema-ref-parser";
import { compile } from "json-schema-to-typescript";

const root = process.cwd();
const schemaDir = path.join(root, "packages", "protocol", "schema");
const outDir = path.join(root, "packages", "protocol", "typescript");
const outFile = path.join(outDir, "generated.ts");

const entries = [
  {
    file: "edit-action.replace-between-anchors.schema.json",
    name: "ReplaceBetweenAnchorsAction",
  },
  { file: "edit-action.schema.json", name: "EditAction" },
  { file: "ping.params.schema.json", name: "PingParams" },
  { file: "ping.result.schema.json", name: "PingResult" },
  { file: "edit-apply.params.schema.json", name: "EditApplyParams" },
  { file: "edit-apply.result.schema.json", name: "EditApplyResult" },
];

mkdirSync(outDir, { recursive: true });

const compilerOptions = {
  bannerComment: "",
  style: {
    singleQuote: false,
    semi: true,
  },
};

const typeMap = new Map(); // name → definition

function extractTypes(ts) {
  const regex = /export (interface|type) (\w+)[\s\S]*?\n}/g;
  const results = [];
  let match;

  while ((match = regex.exec(ts)) !== null) {
    results.push({
      name: match[2],
      code: match[0],
    });
  }

  return results;
}

for (const entry of entries) {
  const schemaPath = path.join(schemaDir, entry.file);
  const dereferenced = await $RefParser.dereference(schemaPath);

  const ts = await compile(dereferenced, entry.name, compilerOptions);

  const types = extractTypes(ts);

  for (const t of types) {
    if (!typeMap.has(t.name)) {
      typeMap.set(t.name, t.code.trim());
    }
  }
}

const header = "// AUTO-GENERATED. DO NOT EDIT.\n";

const output = [header, ...typeMap.values()].join("\n\n");

writeFileSync(outFile, output, "utf8");

console.log(`Generated ${path.relative(root, outFile)}`);
