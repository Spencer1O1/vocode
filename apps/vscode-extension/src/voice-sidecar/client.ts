import type { ChildProcessWithoutNullStreams } from "node:child_process";
import { createInterface } from "node:readline";

interface VoiceSidecarEvent {
  type: string;
  state?: string;
  message?: string;
  version?: string;
  text?: string;
}

export interface VoiceSidecarTranscriptEvent {
  type: "transcript";
  text: string;
  timestamp?: number;
}

export class VoiceSidecarClient {
  private readonly process: ChildProcessWithoutNullStreams;
  private disposed = false;
  private transcriptHandler?: (evt: VoiceSidecarTranscriptEvent) => void;
  private stateHandler?: (evt: { type: "state"; state: string }) => void;
  private errorHandler?: (evt: { type: "error"; message: string }) => void;

  constructor(process: ChildProcessWithoutNullStreams) {
    this.process = process;

    const rl = createInterface({
      input: this.process.stdout,
      crlfDelay: Number.POSITIVE_INFINITY,
    });

    rl.on("line", (line) => {
      const trimmed = line.trim();
      if (!trimmed) {
        return;
      }
      try {
        const evt = JSON.parse(trimmed) as VoiceSidecarEvent;
        if (evt.type === "ready") {
          console.log(`[vocode-voiced] ready version=${evt.version ?? "?"}`);
        } else if (evt.type === "state") {
          console.log(`[vocode-voiced] state=${evt.state ?? "?"}`);
          if (typeof evt.state === "string" && evt.state) {
            this.stateHandler?.({ type: "state", state: evt.state });
          }
        } else if (evt.type === "error") {
          console.warn(`[vocode-voiced] error: ${evt.message ?? "unknown"}`);
          const message =
            typeof evt.message === "string" ? evt.message : "unknown";
          this.errorHandler?.({ type: "error", message });
        } else if (evt.type === "transcript") {
          const text = typeof evt.text === "string" ? evt.text : "";
          if (!text) return;
          this.transcriptHandler?.({ type: "transcript", text });
        }
      } catch (err) {
        console.error("[vocode-voiced] failed to parse stdout as JSON:", err);
        console.error("[vocode-voiced] raw line:", trimmed);
      }
    });
  }

  public start(): void {
    this.send({ type: "start" });
  }

  public stop(): void {
    this.send({ type: "stop" });
  }

  public shutdown(): void {
    this.send({ type: "shutdown" });
  }

  public dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    this.shutdown();
  }

  public onTranscript(handler: (evt: VoiceSidecarTranscriptEvent) => void) {
    this.transcriptHandler = handler;
  }

  public onState(handler: (evt: { type: "state"; state: string }) => void) {
    this.stateHandler = handler;
  }

  public onError(handler: (evt: { type: "error"; message: string }) => void) {
    this.errorHandler = handler;
  }

  private send(msg: { type: string }): void {
    if (this.disposed) return;
    try {
      this.process.stdin.write(`${JSON.stringify(msg)}\n`);
    } catch (err) {
      console.error("[vocode-voiced] failed to write stdin:", err);
    }
  }
}
