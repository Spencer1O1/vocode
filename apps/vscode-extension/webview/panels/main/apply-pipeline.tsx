import type { PendingRow } from "../../types";

export type ApplyStepVisual = "done" | "active" | "pending";

export function applyPipelineSteps(status: PendingRow["status"]): {
  label: string;
  visual: ApplyStepVisual;
  title?: string;
}[] {
  const st = status;
  return [
    { label: "Transcript committed", visual: "done" },
    {
      label: "Run agent",
      visual: st === "processing" ? "active" : "pending",
      title:
        st === "queued"
          ? "Waiting to send this line to the daemon"
          : "Agent loop is running",
    },
  ];
}

export function ApplyStepRow({
  label,
  visual,
  title,
}: {
  label: string;
  visual: ApplyStepVisual;
  title?: string;
}) {
  return (
    <div className={`apply-step apply-step-${visual}`} title={title}>
      <span className="apply-step-mark" aria-hidden="true" />
      <span className="apply-step-label">{label}</span>
    </div>
  );
}
