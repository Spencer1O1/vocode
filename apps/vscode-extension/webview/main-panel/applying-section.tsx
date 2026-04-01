import type { PendingRow } from "../types";
import { fmtTime, statusBadgeTitle, statusLabel } from "../util";
import { ApplyStepRow, applyPipelineSteps } from "./apply-pipeline";
import { CompactQueuedCard } from "./compact-queued-card";

export function ApplyingSection({
  pending,
}: {
  pending: readonly PendingRow[];
}) {
  const primary = pending[0];
  const queuedRest = pending.length > 1 ? pending.slice(1) : [];

  return (
    <section className="panel-section">
      <h1>Applying</h1>
      {primary ? (
        <div className="stack applying-stack">
          <div
            className={`card pending applying-primary-card ${primary.status}`}
          >
            <div className="meta">
              <span
                className="badge"
                title={statusBadgeTitle(primary.status) || undefined}
              >
                {statusLabel(primary.status)}
              </span>
              <span>{fmtTime(primary.receivedAt)}</span>
            </div>
            <div className="text">{primary.text}</div>
            <div className="apply-steps" role="list" aria-label="Pipeline">
              {applyPipelineSteps(primary.status).map((s) => (
                <ApplyStepRow key={s.label} {...s} />
              ))}
            </div>
          </div>
          {queuedRest.length > 0 ? (
            <div className="applying-queue-block">
              <h2 className="applying-subhead">Queued ({queuedRest.length})</h2>
              <div className="stack applying-queue-stack">
                {queuedRest.map((p) => (
                  <CompactQueuedCard key={p.id} p={p} />
                ))}
              </div>
            </div>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}
