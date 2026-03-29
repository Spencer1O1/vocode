import type { HandledRow, PanelState, PendingRow } from "./types";
import { fmtTime, statusBadgeTitle, statusLabel } from "./util";

function LiveSection({ state }: { state: PanelState }) {
  const voiceListening = state.voiceListening === true;
  const partial =
    typeof state.latestPartial === "string" && state.latestPartial.length > 0
      ? state.latestPartial
      : null;
  const showLive = voiceListening && partial !== null;

  return (
    <>
      <h1>Live</h1>
      {!voiceListening ? (
        <div className="empty" />
      ) : !showLive ? (
        <div className="empty" />
      ) : (
        <div className="stack">
          <div className="card live">
            <div className="meta">
              <span
                className="badge"
                title="Streaming speech-to-text — not final until it moves to Done"
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
      )}
    </>
  );
}

function PendingCard({ p }: { p: PendingRow }) {
  const bt = statusBadgeTitle(p.status);
  return (
    <div className={`card pending ${p.status}`}>
      <div className="meta">
        <span className="badge" title={bt || undefined}>
          {statusLabel(p.status)}
        </span>
        <span>{fmtTime(p.receivedAt)}</span>
      </div>
      <div className="text">{p.text}</div>
    </div>
  );
}

function ApplyingSection({ pending }: { pending: readonly PendingRow[] }) {
  return (
    <>
      <h1 className="section-title">Applying</h1>
      {!pending.length ? (
        <div className="empty" />
      ) : (
        <div className="stack">
          {pending.map((p) => (
            <PendingCard key={p.id} p={p} />
          ))}
        </div>
      )}
    </>
  );
}

function DoneCard({ h }: { h: HandledRow }) {
  const failed =
    typeof h.errorMessage === "string" && h.errorMessage.length > 0;
  const cardCls = failed ? "card done failed" : "card done";
  return (
    <div className={cardCls}>
      <div className="meta">
        {failed ? (
          <span
            className="badge"
            title="Daemon or workspace apply did not succeed"
          >
            {"Couldn't run"}
          </span>
        ) : null}
        <span>{fmtTime(h.receivedAt)}</span>
      </div>
      <div className="text">{h.text}</div>
      {failed ? (
        <div className="error-detail">Error: {h.errorMessage}</div>
      ) : null}
    </div>
  );
}

function DoneSection({
  recentHandled,
}: {
  recentHandled: readonly HandledRow[];
}) {
  return (
    <>
      <h1 className="section-title">Done</h1>
      {!recentHandled.length ? (
        <div className="empty" />
      ) : (
        <div className="stack">
          {recentHandled.map((h) => (
            <DoneCard key={`${h.receivedAt}-done`} h={h} />
          ))}
        </div>
      )}
    </>
  );
}

function SummarySection({
  recentHandled,
}: {
  recentHandled: readonly HandledRow[];
}) {
  const withSummaries = recentHandled.filter(
    (h) => typeof h.summary === "string" && h.summary.length > 0,
  );
  return (
    <>
      <h1 className="section-title">Summary</h1>
      {!withSummaries.length ? (
        <div className="empty" />
      ) : (
        <div className="stack">
          {withSummaries.map((h) => {
            const preview =
              typeof h.text === "string" && h.text.length > 140
                ? `${h.text.slice(0, 140)}…`
                : h.text || "";
            return (
              <div key={`${h.receivedAt}-summary`} className="card summary">
                <div className="meta">
                  <span
                    className="badge"
                    title="Agent done summary for this turn"
                  >
                    Summary
                  </span>
                  <span>{fmtTime(h.receivedAt)}</span>
                </div>
                <div className="text">{h.summary}</div>
                {preview ? (
                  <div className="summary-for">Transcript: {preview}</div>
                ) : null}
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}

export function MainPanel({ state }: { state: PanelState }) {
  const pending = Array.isArray(state.pending) ? state.pending : [];
  const recentHandled = Array.isArray(state.recentHandled)
    ? state.recentHandled
    : [];

  return (
    <div id="main-root">
      <LiveSection state={state} />
      <ApplyingSection pending={pending} />
      <DoneSection recentHandled={recentHandled} />
      <SummarySection recentHandled={recentHandled} />
      <p className="hint">Vocode · voice to code</p>
    </div>
  );
}
