import { spawnSync } from "node:child_process";
import { copyFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { FAT_TARGETS } from "./vsix-fat-targets.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(scriptDir, "..", "..");
const extBinRoot = path.join(repoRoot, "apps", "vscode-extension", "bin");
const coreDir = path.join(repoRoot, "apps", "core");
const goCache = path.join(coreDir, ".gocache");
mkdirSync(goCache, { recursive: true });

/** @type {Map<string, { goos: string; goarch: string; goarm?: string }>} */
const uniqueGo = new Map();
for (const t of FAT_TARGETS) {
  const key = `${t.goos}/${t.goarch}${t.goarm ? `v${t.goarm}` : ""}`;
  if (!uniqueGo.has(key)) {
    uniqueGo.set(key, { goos: t.goos, goarch: t.goarch, goarm: t.goarm });
  }
}

const tmpDir = path.join(coreDir, ".cross-build-tmp");
mkdirSync(tmpDir, { recursive: true });

for (const [key, g] of uniqueGo) {
  const win = g.goos === "windows";
  const binName = win ? "vocode-cored.exe" : "vocode-cored";
  const tmpOut = path.join(tmpDir, key.replace(/[/]/g, "_"), binName);
  mkdirSync(path.dirname(tmpOut), { recursive: true });

  const env = {
    ...process.env,
    GOOS: g.goos,
    GOARCH: g.goarch,
    CGO_ENABLED: "0",
    GOCACHE: goCache,
  };
  if (g.goarm) {
    env.GOARM = g.goarm;
  }

  console.log(
    `[build-core-cross] GOOS=${g.goos} GOARCH=${g.goarch}${g.goarm ? ` GOARM=${g.goarm}` : ""}`,
  );
  const r = spawnSync(
    "go",
    [
      "build",
      "-buildvcs=false",
      "-trimpath",
      "-o",
      tmpOut,
      "./cmd/vocode-cored",
    ],
    { cwd: coreDir, env, stdio: "inherit" },
  );
  if (r.status !== 0) {
    process.exit(r.status ?? 1);
  }

  for (const t of FAT_TARGETS) {
    if (t.goos !== g.goos || t.goarch !== g.goarch) continue;
    if ((t.goarm || "") !== (g.goarm || "")) continue;
    const destDir = path.join(extBinRoot, t.slug);
    mkdirSync(destDir, { recursive: true });
    const dest = path.join(destDir, binName);
    copyFileSync(tmpOut, dest);
    console.log(`[build-core-cross] -> ${path.relative(repoRoot, dest)}`);
  }
}

console.log("[build-core-cross] done");
