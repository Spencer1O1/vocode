import { getVsCodeApi } from "../../api/vscode";
import type { PanelState } from "../../types";

export function ClarifyPanel({ state }: { state: PanelState }) {
  const prompt = state.clarifyPrompt;
  if (!prompt?.question) {
    return null;
  }
  return (
    <div className="interrupt-panel">
      <p className="interrupt-panel-kicker">
        Vocode needs a detail before it can continue.
      </p>
      <div className="card interrupt-panel-card">
        <div className="meta">
          <span className="badge" title="Answer with your voice next">
            Question
          </span>
        </div>
        <p className="interrupt-panel-lead">{prompt.question}</p>
        <div className="interrupt-panel-divider" />
        <p className="interrupt-panel-label">Your last instruction</p>
        <div className="text interrupt-panel-transcript">
          {prompt.originalTranscript}
        </div>
      </div>
      <p className="interrupt-panel-footnote">
        The next committed utterance is sent as your reply to this question.
      </p>
      <div className="interrupt-actions">
        <button
          type="button"
          className="interrupt-secondary-btn"
          onClick={() =>
            getVsCodeApi()?.postMessage({
              type: "transcriptControl",
              control: "cancel_clarify",
            })
          }
        >
          Cancel — skip this transcript
        </button>
      </div>
    </div>
  );
}
