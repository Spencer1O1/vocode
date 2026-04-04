/**
 * Download ripgrep for every VSIX fat target into tools/ripgrep/<slug>/.
 * Fetches microsoft/ripgrep-prebuilt releases. Archive type follows the *target*
 * (zip for win32, tar.gz for darwin/linux), not the host OS — required when
 * provisioning from Windows.
 */

import { spawnSync } from "node:child_process";
import {
  copyFileSync,
  createWriteStream,
  existsSync,
  mkdirSync,
  readdirSync,
  rmSync,
  statSync,
} from "node:fs";
import path from "node:path";
import { pipeline } from "node:stream/promises";
import { fileURLToPath } from "node:url";

import { FAT_TARGETS } from "./vsix-fat-targets.mjs";

const REPO = "microsoft/ripgrep-prebuilt";
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(scriptDir, "..", "..");

const tmpBase = path.join(repoRoot, "tools", ".ripgrep-fat-cache");
const downloadDir = path.join(tmpBase, "_downloads");
mkdirSync(downloadDir, { recursive: true });

/**
 * @param {string} url
 * @param {Record<string, string>} headers
 */
async function fetchJson(url, headers) {
  const r = await fetch(url, { headers });
  if (!r.ok) {
    throw new Error(`GET ${url} → ${r.status}`);
  }
  return r.json();
}

/**
 * @param {string} url
 * @param {string} destPath
 * @param {Record<string, string>} headers
 */
async function downloadToFile(url, destPath, headers) {
  const r = await fetch(url, { headers });
  if (!r.ok) {
    throw new Error(`Download ${url} → ${r.status}`);
  }
  await pipeline(r.body, createWriteStream(destPath));
}

/**
 * @param {string} p
 */
function rmrf(p) {
  rmSync(p, { recursive: true, force: true });
}

/**
 * @param {string} archivePath
 * @param {string} destDir
 */
function extractArchive(archivePath, destDir) {
  rmrf(destDir);
  mkdirSync(destDir, { recursive: true });
  const r = spawnSync("tar", ["xf", archivePath, "-C", destDir], {
    stdio: "inherit",
  });
  if (r.error) {
    throw r.error;
  }
  if (r.status !== 0) {
    throw new Error(`tar xf exited with ${r.status}`);
  }
}

/**
 * @param {{ version: string; target: string; slug: string; force: boolean }} opts
 */
async function ensureRipgrepBinary(opts) {
  const useZip = opts.slug.startsWith("win32");
  const ext = useZip ? ".zip" : ".tar.gz";
  const assetName = ["ripgrep", opts.version, opts.target].join("-") + ext;
  const archivePath = path.join(downloadDir, assetName);

  const hdr = { "user-agent": "vocode-provision-ripgrep" };
  if (process.env.GITHUB_TOKEN) {
    hdr.authorization = `token ${process.env.GITHUB_TOKEN}`;
  }

  if (!opts.force && existsSync(archivePath)) {
    console.log(`[provision-ripgrep-all] Using cached ${assetName}`);
  } else {
    const apiUrl = `https://api.github.com/repos/${REPO}/releases/tags/${opts.version}`;
    console.log(`GET ${apiUrl}`);
    const release = await fetchJson(apiUrl, hdr);
    const asset = release.assets?.find((a) => a.name === assetName);
    if (!asset) {
      throw new Error(`Asset not found: ${assetName}`);
    }
    const dlUrl = asset.browser_download_url;
    console.log(`Downloading ${dlUrl}`);
    await downloadToFile(dlUrl, archivePath, hdr);
  }

  const tmpUnpack = path.join(tmpBase, opts.slug);
  extractArchive(archivePath, tmpUnpack);

  let found = path.join(tmpUnpack, "rg");
  if (!existsSync(found)) {
    found = path.join(tmpUnpack, "rg.exe");
  }
  if (!existsSync(found)) {
    const walk = (d) => {
      for (const name of readdirSync(d)) {
        const p = path.join(d, name);
        const st = statSync(p);
        if (st.isDirectory()) {
          const w = walk(p);
          if (w) return w;
        } else if (name === "rg" || name === "rg.exe") {
          return p;
        }
      }
      return null;
    };
    found = walk(tmpUnpack);
  }
  if (!found) {
    throw new Error(`rg not found under ${tmpUnpack}`);
  }
  return found;
}

for (const t of FAT_TARGETS) {
  const destDir = path.join(repoRoot, "tools", "ripgrep", t.slug);
  const win = t.slug.startsWith("win32");
  const rgName = win ? "rg.exe" : "rg";
  const destRg = path.join(destDir, rgName);

  console.log(
    `[provision-ripgrep-all] ${t.slug} (${t.rgTarget} @ ${t.rgVersion})`,
  );
  try {
    const found = await ensureRipgrepBinary({
      version: t.rgVersion,
      target: t.rgTarget,
      slug: t.slug,
      force: false,
    });
    mkdirSync(destDir, { recursive: true });
    copyFileSync(found, destRg);
    console.log(`[provision-ripgrep-all] -> tools/ripgrep/${t.slug}/${rgName}`);
  } catch (e) {
    console.error(`[provision-ripgrep-all] FAILED ${t.slug}:`, e?.message || e);
    process.exit(1);
  }
}

console.log("[provision-ripgrep-all] done");
