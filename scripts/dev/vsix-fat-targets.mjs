import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const jsonPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "vsix-fat-targets.json",
);

/** @type {{ slug: string; goos: string; goarch: string; goarm?: string; rgTarget: string; rgVersion: string }[]} */
export const FAT_TARGETS = JSON.parse(readFileSync(jsonPath, "utf8"));
