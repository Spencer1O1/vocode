/**
 * Native PortAudio build for the current machine only (one VS Code triple).
 * Invoked from apps/voice via: pnpm --filter @vocode/voice build
 */
import {
  buildVoiceForSlug,
  defaultRepoRoot,
  getHostSlug,
} from "./voice-build-lib.mjs";

const repoRoot = defaultRepoRoot();
const slug = getHostSlug();

buildVoiceForSlug(slug, { repoRoot });
console.log(`[build-voice] done (${slug})`);
