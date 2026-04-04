"use strict";

/**
 * vocode-cored is built into apps/vscode-extension/bin/<platform-arch>/ by build-core.mjs (host)
 * or build-core-cross.mjs (fat VSIX).
 * Merges vocode-voiced into each platform bin folder, copies ripgrep, LICENSE, embeds protocol for the VSIX.
 *
 * Slim (default): single host triple from process.platform-arch.
 * Fat (--fat): every slug in scripts/dev/vsix-fat-targets.json (cross-platform universal VSIX).
 */

const fs = require("node:fs");
const path = require("node:path");

const extRoot = path.join(__dirname, "..");
const repoRoot = path.join(extRoot, "..", "..");
const target = `${process.platform}-${process.arch}`;
const fat = process.argv.includes("--fat");

const fatTargetsPath = path.join(
  repoRoot,
  "scripts",
  "dev",
  "vsix-fat-targets.json",
);

function rmrf(p) {
  fs.rmSync(p, { recursive: true, force: true });
}

function copyDir(src, dest) {
  if (!fs.existsSync(src)) {
    return false;
  }
  fs.mkdirSync(path.dirname(dest), { recursive: true });
  fs.cpSync(src, dest, { recursive: true });
  return true;
}

function stageProtocolAndLicense() {
  const licenseSrc = path.join(repoRoot, "LICENSE");
  const licenseDest = path.join(extRoot, "LICENSE");
  if (fs.existsSync(licenseSrc)) {
    fs.copyFileSync(licenseSrc, licenseDest);
  }

  const protoDist = path.join(repoRoot, "packages", "protocol", "dist");
  if (!fs.existsSync(protoDist)) {
    console.error(
      `[stage-marketplace-assets] Missing ${protoDist}\n` +
        "  Run: pnpm --filter @vocode/protocol build",
    );
    process.exit(1);
  }
  const protoPkg = path.join(extRoot, "dist", "protocol-pkg");
  fs.rmSync(protoPkg, { recursive: true, force: true });
  fs.mkdirSync(protoPkg, { recursive: true });
  fs.cpSync(protoDist, path.join(protoPkg, "dist"), { recursive: true });
  fs.writeFileSync(
    path.join(protoPkg, "package.json"),
    `${JSON.stringify(
      {
        name: "@vocode/protocol",
        version: "0.0.0",
        type: "commonjs",
        main: "./dist/index.js",
        types: "./dist/index.d.ts",
      },
      null,
      2,
    )}\n`,
  );

  const clientJs = path.join(extRoot, "dist", "daemon", "client.js");
  if (fs.existsSync(clientJs)) {
    const code = fs.readFileSync(clientJs, "utf8");
    const needle = 'require("@vocode/protocol")';
    const replacement = 'require("../protocol-pkg")';
    if (!code.includes(needle)) {
      if (!code.includes(replacement)) {
        console.warn(
          `[stage-marketplace-assets] Expected ${needle} in daemon/client.js — skip protocol rewrite`,
        );
      }
    } else {
      fs.writeFileSync(clientJs, code.split(needle).join(replacement));
    }
  }
}

rmrf(path.join(extRoot, "tools", "ripgrep"));

if (fat) {
  if (!fs.existsSync(fatTargetsPath)) {
    console.error(`[stage-marketplace-assets] Missing ${fatTargetsPath}`);
    process.exit(1);
  }
  /** @type {{ slug: string }[]} */
  const fatTargets = JSON.parse(fs.readFileSync(fatTargetsPath, "utf8"));

  for (const t of fatTargets) {
    const slug = t.slug;
    const win = slug.startsWith("win32");
    const coreBinary = win ? "vocode-cored.exe" : "vocode-cored";
    const corePath = path.join(extRoot, "bin", slug, coreBinary);
    if (!fs.existsSync(corePath)) {
      console.error(
        `[stage-marketplace-assets] Missing core daemon for ${slug}: ${corePath}\n` +
          "  Run from repo root: node scripts/dev/build-core-cross.mjs",
      );
      process.exit(1);
    }

    const voiceBinDir = path.join(repoRoot, "apps", "voice", "bin", slug);
    const voiceName = win ? "vocode-voiced.exe" : "vocode-voiced";
    const voicePath = path.join(voiceBinDir, voiceName);
    if (!fs.existsSync(voicePath)) {
      console.error(
        `[stage-marketplace-assets] Fat VSIX requires voice sidecar for ${slug}. Missing:\n` +
          `  ${voicePath}\n` +
          "  Run from repo root: node scripts/dev/build-voice-cross.mjs (included in pnpm vscode:package:fat).",
      );
      process.exit(1);
    }
    const outBinDir = path.join(extRoot, "bin", slug);
    copyDir(voiceBinDir, outBinDir);

    const rgDir = path.join(repoRoot, "tools", "ripgrep", slug);
    const outRgDir = path.join(extRoot, "tools", "ripgrep", slug);
    if (!copyDir(rgDir, outRgDir)) {
      console.error(
        `[stage-marketplace-assets] Missing ripgrep for ${slug}: ${rgDir}\n` +
          "  Run from repo root: node scripts/dev/provision-ripgrep-all.mjs",
      );
      process.exit(1);
    }
  }

  stageProtocolAndLicense();
  console.log(
    `[stage-marketplace-assets] Fat VSIX: staged ${fatTargets.length} platform folders → ${extRoot}`,
  );
  process.exit(0);
}

const voiceBinDir = path.join(repoRoot, "apps", "voice", "bin", target);
const rgDir = path.join(repoRoot, "tools", "ripgrep", target);
const outBinDir = path.join(extRoot, "bin", target);
const outRgDir = path.join(extRoot, "tools", "ripgrep", target);
const coreBinary =
  process.platform === "win32" ? "vocode-cored.exe" : "vocode-cored";
const corePath = path.join(outBinDir, coreBinary);

if (!fs.existsSync(corePath)) {
  console.error(
    `[stage-marketplace-assets] Missing core daemon at ${corePath}\n` +
      "  Run: pnpm --filter @vocode/core build",
  );
  process.exit(1);
}

if (!copyDir(voiceBinDir, outBinDir)) {
  console.warn(
    `[stage-marketplace-assets] No voice sidecar at ${voiceBinDir} (voice features will not work in this VSIX until you build it).`,
  );
}

if (!copyDir(rgDir, outRgDir)) {
  console.error(
    `[stage-marketplace-assets] Missing ripgrep at ${rgDir}\n` +
      "  Run from repo root: pnpm provision:ripgrep",
  );
  process.exit(1);
}

stageProtocolAndLicense();

console.log(
  `[stage-marketplace-assets] Staged bin + tools for ${target} → ${extRoot}`,
);
