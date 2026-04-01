import { LivePartialCard } from "../../live-partial-card";
import type { PanelState } from "../../types";

export function LiveSection({ state }: { state: PanelState }) {
  return (
    <section className="panel-section">
      <h1>Live</h1>
      <LivePartialCard state={state} />
    </section>
  );
}
