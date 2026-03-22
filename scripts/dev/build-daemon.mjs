import { spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import path from "node:path";

const target = `${process.platform}-${process.arch}`;
const binary = process.platform === "win32" ? "vocoded.exe" : "vocoded";

mkdirSync(path.join("bin", target), { recursive: true });

const result = spawnSync(
  "go",
  ["build", "-o", path.join("bin", target, binary), "./cmd/vocoded"],
  {
    stdio: "inherit",
  },
);

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}
