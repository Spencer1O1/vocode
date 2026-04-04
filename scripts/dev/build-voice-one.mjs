/**
 * Build vocode-voiced with PortAudio for a single VSIX slug (for CI matrix jobs).
 * Usage: node scripts/dev/build-voice-one.mjs darwin-arm64
 *    or: VOCODE_VOICE_SLUG=darwin-arm64 node scripts/dev/build-voice-one.mjs
 */
import { buildVoiceForSlug, defaultRepoRoot } from "./voice-build-lib.mjs";

const slug = process.argv[2] || process.env.VOCODE_VOICE_SLUG;
if (!slug) {
  console.error(
    "Usage: node scripts/dev/build-voice-one.mjs <slug>\n" +
      "Example: node scripts/dev/build-voice-one.mjs linux-arm64",
  );
  process.exit(1);
}

buildVoiceForSlug(slug, { repoRoot: defaultRepoRoot() });
console.log(`[build-voice-one] done (${slug})`);
