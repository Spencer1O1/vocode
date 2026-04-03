import type { PendingRow } from "../types";
import { fmtTime, statusBadgeTitle, statusLabel } from "../util";
import {
  ProcessingStepRow,
  processingPipelineSteps,
} from "./main/processing-pipeline";

/**
 * Compact “Applying” block for interrupt views (search / clarify) so users still see
 * which transcript is running without returning to the main panel.
 */
export function ProcessingStrip({
  pending,
}: {
  pending: readonly PendingRow[];
}) {
  const p = pending[0];
  if (!p) {
    return null;
  }
  return (
    <div className="pending-apply-strip" aria-live="polite">
      <div className="pending-apply-strip-head">
        <span className="badge" title={statusBadgeTitle(p.status) || undefined}>
          {statusLabel(p.status)}
        </span>
        <span className="pending-apply-strip-time">
          {fmtTime(p.receivedAt)}
        </span>
      </div>
      <div className="pending-apply-strip-text">{p.text}</div>
      <div
        className="pending-apply-strip-steps"
        role="list"
        aria-label="Pipeline"
      >
        {processingPipelineSteps(p.status).map((s) => (
          <ProcessingStepRow key={s.label} {...s} />
        ))}
      </div>
    </div>
  );
}
