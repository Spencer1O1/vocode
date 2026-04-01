import type { PanelState } from "../types";

export function ClarifySection({ state }: { state: PanelState }) {
  const prompt = state.clarifyPrompt;
  if (!prompt || !prompt.question) {
    return null;
  }
  return (
    <section className="panel-section">
      <h1>Clarification needed</h1>
      <div className="stack">
        <div className="card done failed history-card">
          <div className="meta">
            <span className="badge" title="Vocode needs one detail to proceed">
              Question
            </span>
          </div>
          <div className="text history-summary">{prompt.question}</div>
          <div className="history-transcript muted-transcript">
            Original: {prompt.originalTranscript}
          </div>
          <div className="hint">
            Speak your answer next — Vocode will treat the next utterance as the
            reply.
          </div>
        </div>
      </div>
    </section>
  );
}
