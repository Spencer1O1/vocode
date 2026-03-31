import type { ChildProcessWithoutNullStreams } from "node:child_process";
import type {
  PingParams,
  PingResult,
  VoiceTranscriptCompletion,
  VoiceTranscriptParams,
} from "@vocode/protocol";
import { isPingResult, isVoiceTranscriptCompletion } from "@vocode/protocol";

import { RpcTransport } from "./rpc-transport";

export class DaemonClient {
  private readonly transport: RpcTransport;

  constructor(process: ChildProcessWithoutNullStreams) {
    this.transport = new RpcTransport(process);
  }

  public registerRequestHandler(
    method: string,
    handler: (params: unknown) => Promise<unknown> | unknown,
  ): void {
    this.transport.registerRequestHandler(method, handler);
  }

  public async sendRequest<T>(
    method: string,
    params: unknown,
    isResult?: (value: unknown) => value is T,
  ): Promise<T> {
    const result = await this.transport.request(method, params);

    if (isResult && !isResult(result)) {
      throw new Error(`Invalid ${method} response from daemon.`);
    }

    return result as T;
  }

  public ping(params: PingParams = {}): Promise<PingResult> {
    return this.sendRequest<PingResult>("ping", params, isPingResult);
  }

  public transcript(
    params: VoiceTranscriptParams,
  ): Promise<VoiceTranscriptCompletion> {
    return this.sendRequest<VoiceTranscriptCompletion>(
      "voice.transcript",
      params,
      isVoiceTranscriptCompletion,
    );
  }

  public dispose(): void {
    this.transport.dispose();
  }
}
