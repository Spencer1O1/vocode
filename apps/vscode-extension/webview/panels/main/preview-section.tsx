import { getVsCodeApi } from "../../api/vscode";
import type { PanelState } from "../../types";

export function PreviewSection({
  state,
}: {
  state: PanelState;
}) {
  const preview = state.pendingPreview;
  if (!preview || preview.paths.length === 0) {
    return null;
  }

  const fileCount = preview.paths.length;
  const label = fileCount === 1 ? "1 file" : `${fileCount} files`;

  return (
    <section className="panel-section preview-section">
      <h1>Preview</h1>
      <div className="card preview-card">
        <p className="preview-desc">
          {label} edited — review the changes in the editor, then accept or
          reject.
        </p>
        <div className="preview-actions">
          <button
            type="button"
            className="preview-btn preview-btn-accept"
            onClick={() =>
              getVsCodeApi()?.postMessage({ type: "acceptPreview" })
            }
          >
            Accept
          </button>
          <button
            type="button"
            className="preview-btn preview-btn-reject"
            onClick={() =>
              getVsCodeApi()?.postMessage({ type: "rejectPreview" })
            }
          >
            Reject
          </button>
        </div>
      </div>
    </section>
  );
}
