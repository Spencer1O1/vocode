declare global {
  interface Window {
    acquireVsCodeApi?: () => VsCodeApi;
  }
}

export type VsCodeApi = {
  postMessage(message: unknown): void;
  getState(): unknown;
  setState(state: unknown): void;
};

let cached: VsCodeApi | undefined;

/** VS Code injects this once per webview session; absent when not running inside a webview. */
export function getVsCodeApi(): VsCodeApi | undefined {
  if (cached !== undefined) {
    return cached;
  }
  const fn =
    typeof window !== "undefined" ? window.acquireVsCodeApi : undefined;
  if (typeof fn !== "function") {
    return undefined;
  }
  cached = fn();
  return cached;
}
