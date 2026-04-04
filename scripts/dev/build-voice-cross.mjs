/**
 * Build vocode-voiced with PortAudio for every VSIX fat target (or a CI subset).
 *
 * Full local run (from repo root): builds
 *   - linux-* : Docker (Debian glibc) except native when host Linux matches slug
 *   - alpine-* : Docker (Alpine musl)
 *   - win32-* : native Windows only when slug matches host (x64 vs arm64)
 *   - darwin-* : native macOS only when slug matches host (or cross-arch on Mac)
 *
 * Requires Docker (Desktop) for Linux/Alpine when not on matching Linux host.
 *
 * Outside CI, if a slug cannot be built on this machine but `apps/voice/bin/<slug>/`
 * already has a binary (e.g. copied from another OS job), that slug is kept and a
 * warning is printed. CI (GITHUB_ACTIONS) always fails the job if any build fails.
 *
 * CI subsets (parallel jobs):
 *   VOCODE_VOICE_BUILD_FAMILY=linux   → linux-* + alpine-* only
 *   VOCODE_VOICE_BUILD_FAMILY=windows → win32-* for current Windows arch only
 *   VOCODE_VOICE_BUILD_FAMILY=darwin  → darwin-* for current Mac arch only
 *
 * Single slug: VOCODE_VOICE_SLUG=linux-x64 or first argv.
 */
import { existsSync } from "node:fs";
import path from "node:path";

import {
  buildVoiceForSlug,
  defaultRepoRoot,
  slugsForFamily,
} from "./voice-build-lib.mjs";

const repoRoot = defaultRepoRoot();
const inCi = process.env.CI === "true" || process.env.GITHUB_ACTIONS === "true";

function voiceOutputPath(slug) {
  const win = slug.startsWith("win32");
  const name = win ? "vocode-voiced.exe" : "vocode-voiced";
  return path.join(repoRoot, "apps", "voice", "bin", slug, name);
}

const familyRaw = process.env.VOCODE_VOICE_BUILD_FAMILY;
/** @type {'linux' | 'windows' | 'darwin' | null} */
const family =
  familyRaw === "linux" || familyRaw === "windows" || familyRaw === "darwin"
    ? familyRaw
    : null;

const singleFromEnv = process.env.VOCODE_VOICE_SLUG;
const singleFromArgv = process.argv[2]?.startsWith("-")
  ? null
  : process.argv[2];
const singleSlug = singleFromEnv || singleFromArgv;

let slugs;
if (singleSlug) {
  slugs = [singleSlug];
} else {
  slugs = slugsForFamily(family, repoRoot);
}

console.log(
  `[build-voice-cross] ${slugs.length} slug(s): ${slugs.join(", ")}${family ? ` (family=${family})` : ""}`,
);

for (const slug of slugs) {
  try {
    buildVoiceForSlug(slug, { repoRoot });
  } catch (err) {
    const out = voiceOutputPath(slug);
    if (!inCi && !singleSlug && existsSync(out)) {
      console.warn(
        `[build-voice-cross] keeping existing ${slug} (${err?.message || err})`,
      );
      continue;
    }
    console.error(err);
    process.exit(1);
  }
}

console.log("[build-voice-cross] done");
