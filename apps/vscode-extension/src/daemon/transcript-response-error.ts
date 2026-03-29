const staleDaemonHint =
  "Daemon/extension protocol mismatch detected (received legacy {accepted:boolean} shape). If you pulled new changes, run: pnpm codegen && pnpm --filter @vocode/daemon build && pnpm --filter @vocode/vscode-extension build, then reload the Extension Development Host.";

export function voiceTranscriptResponseValidationError(result: unknown): string {
  if (
    typeof result === "object" &&
    result !== null &&
    "accepted" in result &&
    typeof (result as { accepted?: unknown }).accepted === "boolean"
  ) {
    return staleDaemonHint;
  }

  return "Invalid voice.transcript response from daemon.";
}
