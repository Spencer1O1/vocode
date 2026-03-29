import { getVsCodeApi } from "./vscode-api";

export type PanelConfig = {
  voiceVadDebug: boolean;
  voiceSidecarLogProtocol: boolean;
};

function ToggleRow(props: {
  id: string;
  label: string;
  description: string;
  checked: boolean;
  disabled?: boolean;
  onChange: (next: boolean) => void;
}) {
  const { id, label, description, checked, disabled, onChange } = props;
  return (
    <label className="settings-row" htmlFor={id}>
      <div className="settings-row-text">
        <span className="settings-row-label">{label}</span>
        <span className="settings-row-desc">{description}</span>
      </div>
      <input
        id={id}
        type="checkbox"
        className="settings-toggle"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
    </label>
  );
}

export function SettingsPanel(props: { config: PanelConfig | null }) {
  const { config } = props;
  const api = getVsCodeApi();
  const disabled = !api || config === null;

  const patch = (partial: Partial<PanelConfig>) => {
    api?.postMessage({ type: "setPanelConfig", patch: partial });
  };

  return (
    <div className="settings-root">
      {config === null ? (
        <p className="settings-loading">Loading options…</p>
      ) : null}
      <p className="settings-intro">
        Extension options for the voice sidecar. Daemon and STT environment
        variables are still configured in your <code>.env</code> for local dev.
        Toggle changes apply the next time you start the voice sidecar (Stop
        Voice, then Start Voice).
      </p>
      <div className="settings-stack">
        <ToggleRow
          id="vocode-voice-vad-debug"
          label="Voice VAD debug"
          description="Forward VOCODE_VOICE_VAD_DEBUG to the sidecar (verbose [vocode-vad] stderr)."
          checked={config?.voiceVadDebug === true}
          disabled={disabled}
          onChange={(voiceVadDebug) => patch({ voiceVadDebug })}
        />
        <ToggleRow
          id="vocode-voice-protocol-log"
          label="Log sidecar protocol"
          description="Log every JSON line from the voice sidecar to Developer Tools (very noisy)."
          checked={config?.voiceSidecarLogProtocol === true}
          disabled={disabled}
          onChange={(voiceSidecarLogProtocol) =>
            patch({ voiceSidecarLogProtocol })
          }
        />
      </div>
      <button
        type="button"
        className="settings-open-vscode"
        disabled={!api}
        onClick={() => api?.postMessage({ type: "openExtensionSettings" })}
      >
        VS Code Settings
      </button>
    </div>
  );
}
