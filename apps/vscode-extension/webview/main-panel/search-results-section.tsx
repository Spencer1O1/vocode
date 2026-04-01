import type { PanelState } from "../types";

export function SearchResultsSection({ state }: { state: PanelState }) {
  const ss = state.searchState;
  if (!ss || !Array.isArray(ss.results) || ss.results.length === 0) {
    return null;
  }
  const active = Math.min(
    Math.max(0, Number.isFinite(ss.activeIndex) ? ss.activeIndex : 0),
    ss.results.length - 1,
  );
  return (
    <section className="panel-section">
      <h1>Search results</h1>
      <div className="stack">
        {ss.results.map((r, i) => (
          <div
            key={`sr-${r.path}:${r.line}:${r.character}`}
            className={`card history-card ${i === active ? "card-active" : ""}`}
          >
            <div className="meta">
              <span className="badge" title="Result number for voice selection">
                {i + 1}
              </span>
              <span className="muted-transcript">
                {r.path}:{r.line + 1}:{r.character + 1}
              </span>
            </div>
            <div className="text mono">{r.preview}</div>
            {i === active ? (
              <div className="hint">
                Active. Say “next”, “back”, or a number (e.g. “3”) to jump.
              </div>
            ) : null}
          </div>
        ))}
      </div>
    </section>
  );
}
