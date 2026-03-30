import { spawnSync } from "node:child_process";
import { existsSync, mkdirSync } from "node:fs";
import path from "node:path";

// Keep go run happy in shells where %LocalAppData% is missing.
const goCache = path.join(process.cwd(), ".gocache");
mkdirSync(goCache, { recursive: true });

function configurePortAudioCgoEnv() {
  const msysRoot = process.env.MSYS2_ROOT ?? "C:\\tools\\msys64";
  const mingwBin = path.join(msysRoot, "mingw64", "bin");
  const mingwPkgConfigDir = path.join(msysRoot, "mingw64", "lib", "pkgconfig");
  const pkgConfigExe = path.join(mingwBin, "pkg-config.exe");
  const pkgConfExe = path.join(mingwBin, "pkgconf.exe");
  const pkgTool = existsSync(pkgConfigExe)
    ? pkgConfigExe
    : existsSync(pkgConfExe)
      ? pkgConfExe
      : "";
  const gccExe = path.join(mingwBin, "gcc.exe");

  const ok =
    pkgTool !== "" && existsSync(mingwPkgConfigDir) && existsSync(gccExe);
  if (!ok) {
    return {
      ok: false,
      env: {},
      msysRoot,
      mingwBin,
      mingwPkgConfigDir,
    };
  }

  const pathSep = ";";
  const pathPrefix = mingwBin + pathSep;
  const existingPath = process.env.Path ?? process.env.PATH ?? "";

  const env = {
    CGO_ENABLED: "1",
    // Set both PATH and Path so child process sees it reliably on Windows.
    PATH: pathPrefix + existingPath,
    Path: pathPrefix + existingPath,
    PKG_CONFIG_PATH:
      mingwPkgConfigDir +
      (process.env.PKG_CONFIG_PATH
        ? pathSep + process.env.PKG_CONFIG_PATH
        : ""),
    PKG_CONFIG: pkgTool,
    CC: gccExe,
  };

  return { ok: true, env };
}

const portAudioEnv =
  process.platform === "win32"
    ? configurePortAudioCgoEnv()
    : { ok: true, env: {} };

if (process.platform === "win32" && "ok" in portAudioEnv && !portAudioEnv.ok) {
  console.warn(
    `[@vocode/voice] PortAudio cgo toolchain not detected. Expected MSYS2 root at ${portAudioEnv.msysRoot}. ` +
      "Ensure portaudio-2.0.pc is findable via pkg-config and a C compiler (gcc) is available on PATH.",
  );
}

const result = spawnSync("go", ["run", "./cmd/vocode-voiced"], {
  env: {
    ...process.env,
    GOCACHE: goCache,
    // PortAudio mic capture is cgo-only.
    CGO_ENABLED: "1",
    ...(portAudioEnv.ok ? portAudioEnv.env : {}),
  },
  stdio: "inherit",
});

if (result.status !== 0) {
  process.exit(result.status ?? 1);
}
