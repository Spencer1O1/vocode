import { spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import path from "node:path";

const target = `${process.platform}-${process.arch}`;
const binary = process.platform === "win32" ? "vocode-voiced.exe" : "vocode-voiced";

mkdirSync(path.join("bin", target), { recursive: true });

// Keep builds self-contained and reliable across shells/platforms.
const goCache = path.join(process.cwd(), ".gocache");
mkdirSync(goCache, { recursive: true });

const result = spawnSync(
  "go",
  [
    "build",
    "-buildvcs=false",
    "-o",
    path.join("bin", target, binary),
    "./cmd/vocode-voiced",
  ],
  {
    env: {
      ...process.env,
      GOCACHE: goCache,
    },
    stdio: "inherit",
  },
);

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}
