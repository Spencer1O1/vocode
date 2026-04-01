import type { PanelState } from "../types";

export function LiveSection({ state }: { state: PanelState }) {
  const voiceListening = state.voiceListening === true;
  const partial =
    typeof state.latestPartial === "string" && state.latestPartial.length > 0
      ? state.latestPartial
      : null;
  const showLive = voiceListening && partial !== null;

  return (
    <section className="panel-section">
      <h1>Live</h1>
      {showLive ? (
        <div className="stack">
          <div className="card live">
            <div className="meta">
              <span
                className="badge"
                title="Streaming speech-to-text — not final until it moves below"
              >
                Live
              </span>
              <span title="Draft before the provider commits this segment">
                Draft
              </span>
            </div>
            <div className="text">{partial}</div>
            <div className="typing" aria-hidden="true">
              <span className="dot" />
              <span className="dot" />
              <span className="dot" />
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
