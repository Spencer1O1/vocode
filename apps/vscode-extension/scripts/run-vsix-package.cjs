"use strict";

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

if (process.argv.includes("--fat")) {
  process.env.VOCODE_FAT_VSIX = "1";
}

const extRoot = path.join(__dirname, "..");
const pkg = JSON.parse(
  fs.readFileSync(path.join(extRoot, "package.json"), "utf8"),
);
const outDir = path.join(extRoot, "release");
fs.mkdirSync(outDir, { recursive: true });
const outFile = path.join(outDir, `vocode-${pkg.version}.vsix`);

const shell = process.platform === "win32";
const r = spawnSync(
  "pnpm",
  [
    "exec",
    "vsce",
    "package",
    "--no-dependencies",
    "--follow-symlinks",
    "-o",
    outFile,
  ],
  { stdio: "inherit", cwd: extRoot, env: process.env, shell },
);
if (r.error) {
  console.error("[run-vsix-package]", r.error);
  process.exit(1);
}
if (r.status !== 0) {
  process.exit(r.status ?? 1);
}

console.log(`[run-vsix-package] Wrote ${outFile}`);
