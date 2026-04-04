"use strict";

/**
 * Slim (default): protocol + host core + extension build + stage host triple.
 * Fat (VOCODE_FAT_VSIX=1): protocol + cross core + cross voice + extension build + all ripgrep + stage all triples.
 */

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const extRoot = path.join(__dirname, "..");
const repoRoot = path.join(extRoot, "..", "..");
const shell = process.platform === "win32";
const fat = process.env.VOCODE_FAT_VSIX === "1";

function run(cmd, args, cwd) {
  const r = spawnSync(cmd, args, {
    cwd: cwd ?? extRoot,
    stdio: "inherit",
    shell,
    env: process.env,
  });
  if (r.error) {
    console.error("[vscode:prepublish]", r.error);
    process.exit(1);
  }
  if (r.status !== 0) {
    process.exit(r.status ?? 1);
  }
}

run(
  "pnpm",
  ["--dir", repoRoot, "--filter", "@vocode/protocol", "build"],
  repoRoot,
);

if (fat) {
  run(
    process.execPath,
    [path.join(repoRoot, "scripts", "dev", "build-core-cross.mjs")],
    repoRoot,
  );
  run(
    process.execPath,
    [path.join(repoRoot, "scripts", "dev", "build-voice-cross.mjs")],
    repoRoot,
  );
} else {
  run(
    "pnpm",
    ["--dir", repoRoot, "--filter", "@vocode/core", "build"],
    repoRoot,
  );
}

run("pnpm", ["build"], extRoot);

if (fat) {
  run(
    process.execPath,
    [path.join(repoRoot, "scripts", "dev", "provision-ripgrep-all.mjs")],
    repoRoot,
  );
}

const stageArgs = [path.join(__dirname, "stage-marketplace-assets.cjs")];
if (fat) {
  stageArgs.push("--fat");
}
run(process.execPath, stageArgs, extRoot);
