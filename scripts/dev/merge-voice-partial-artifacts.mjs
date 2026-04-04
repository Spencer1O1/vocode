/**
 * Merge CI-downloaded voice artifacts into apps/voice/bin/ (for local fat VSIX).
 *
 * After: gh run download <run-id> -p 'voice-partial-*' -D ./voice-dl
 * Run:   node scripts/dev/merge-voice-partial-artifacts.mjs ./voice-dl
 *
 * Finds .../apps/voice/bin/<slug>/vocode-voiced(.exe) under the download root
 * and copies each slug folder into the repo.
 */
import { cpSync, existsSync, mkdirSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(scriptDir, "..", "..");
const destRoot = path.join(repoRoot, "apps", "voice", "bin");

const root = process.argv[2];
if (!root) {
  console.error(
    "Usage: node scripts/dev/merge-voice-partial-artifacts.mjs <download-root>\n" +
      "Example: gh run download 12345 -p 'voice-partial-*' -D ./dl && node scripts/dev/merge-voice-partial-artifacts.mjs ./dl",
  );
  process.exit(1);
}

const absRoot = path.resolve(root);

/** @param {string} dir */
function walk(dir) {
  /** @type {string[]} */
  const out = [];
  if (!existsSync(dir)) return out;
  for (const name of readdirSync(dir)) {
    const p = path.join(dir, name);
    const st = statSync(p);
    if (st.isDirectory()) {
      out.push(...walk(p));
    } else {
      out.push(p);
    }
  }
  return out;
}

const marker = `${path.sep}apps${path.sep}voice${path.sep}bin${path.sep}`;
const files = walk(absRoot);

/** @type {Map<string, string>} */
const slugToDir = new Map();
for (const file of files) {
  const norm = path.normalize(file);
  const idx = norm.lastIndexOf(marker);
  if (idx === -1) continue;
  const basePath = norm.slice(0, idx + marker.length);
  const rel = norm.slice(idx + marker.length);
  const parts = rel.split(path.sep).filter(Boolean);
  if (parts.length < 2) continue;
  const slug = parts[0];
  const base = parts[1];
  if (!base.startsWith("vocode-voiced")) continue;
  slugToDir.set(slug, path.join(basePath, slug));
}

let n = 0;
for (const [slug, srcDir] of slugToDir) {
  if (!existsSync(srcDir)) continue;
  const destDir = path.join(destRoot, slug);
  mkdirSync(destDir, { recursive: true });
  cpSync(srcDir, destDir, { recursive: true });
  console.log(`[merge-voice] ${slug} <- ${srcDir}`);
  n++;
}

if (n === 0) {
  console.error(
    `[merge-voice] No apps/voice/bin/<slug>/ under ${absRoot} (wrong folder?)`,
  );
  process.exit(1);
}

console.log(`[merge-voice] merged ${n} slug(s) → ${destRoot}`);
