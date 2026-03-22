import type { ChildProcessWithoutNullStreams } from "node:child_process";
import type {
  EditApplyParams,
  EditApplyResult,
  PingParams,
  PingResult,
} from "@vocode/protocol";
import { isEditApplyResult, isPingResult } from "@vocode/protocol";

import { RpcTransport } from "./rpc-transport";

export class DaemonClient {
  private readonly transport: RpcTransport;

  constructor(process: ChildProcessWithoutNullStreams) {
    this.transport = new RpcTransport(process);
  }

  public async ping(params: PingParams = {}): Promise<PingResult> {
    const res = (await this.transport.request("ping", params)) as PingResult;

    if (!isPingResult(res)) {
      throw new Error("Invalid ping response from daemon.");
    }

    return res;
  }

  public async applyEdit(params: EditApplyParams): Promise<EditApplyResult> {
    const res = (await this.transport.request(
      "edit/apply",
      params,
    )) as EditApplyResult;

    if (!isEditApplyResult(res)) {
      throw new Error("Invalid edit/apply response from daemon.");
    }

    return res;
  }

  public dispose(): void {
    this.transport.dispose();
  }
}
