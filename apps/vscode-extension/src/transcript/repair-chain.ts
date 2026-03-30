import type {
  VoiceTranscriptParams,
  VoiceTranscriptResult,
} from "@vocode/protocol";

import type { DaemonClient } from "../daemon/client";
import { applyTranscriptResult } from "./apply-result";
import type { DirectiveApplyOutcome } from "./apply-report-carry";
import {
  clearPendingApplyReport,
  mergeCarriedApplyReportParams,
  carryApplyReportIntoNextRpc,
} from "./apply-report-carry";

let transcriptRepairChainQueue: Promise<void> = Promise.resolve();

type RepairChainResult = {
  lastResult: VoiceTranscriptResult;
  lastOutcomes: DirectiveApplyOutcome[];
  reachedLimit: boolean;
};

async function runRepairChainOnce(args: {
  client: DaemonClient;
  baseParams: VoiceTranscriptParams;
  activeFile: string;
  maxRepairRpcs: number;
}): Promise<RepairChainResult> {
  const { client, baseParams, activeFile, maxRepairRpcs } = args;

  // Ensure leftover apply report queue from a previous run never contaminates this one.
  clearPendingApplyReport();

  let lastResult: VoiceTranscriptResult | undefined;
  let lastOutcomes: DirectiveApplyOutcome[] = [];

  for (let rpcI = 0; rpcI < maxRepairRpcs; rpcI++) {
    const result = await client.transcript(
      mergeCarriedApplyReportParams(baseParams),
    );
    const outcomes = await applyTranscriptResult(result, activeFile);
    carryApplyReportIntoNextRpc(result, outcomes);

    lastResult = result;
    lastOutcomes = outcomes;

    if (!result.success) {
      return {
        lastResult: result,
        lastOutcomes: outcomes,
        reachedLimit: false,
      };
    }

    const dirCount = result.directives?.length ?? 0;
    if (dirCount === 0) {
      return {
        lastResult: result,
        lastOutcomes: outcomes,
        reachedLimit: false,
      };
    }
  }

  if (!lastResult) {
    // Should be impossible: maxRepairRpcs>0 implies at least one call.
    throw new Error("repair-chain: no transcript result");
  }

  return { lastResult, lastOutcomes, reachedLimit: true };
}

/**
 * Queues transcript repair chains so only one committed transcript is ever
 * being applied/auto-repaired at a time.
 */
export function runRepairChainQueued(args: {
  client: DaemonClient;
  baseParams: VoiceTranscriptParams;
  activeFile: string;
  maxRepairRpcs: number;
}): Promise<RepairChainResult> {
  const run = transcriptRepairChainQueue.then(() => runRepairChainOnce(args));
  transcriptRepairChainQueue = run.then(
    () => undefined,
    () => undefined,
  );
  return run;
}
