import type { PanelState } from "./types";

type Props = {
  state: PanelState;
  /**
   * When the mic is on but STT has not produced text yet, still show the live card
   * with a short hint (use on interrupt panels where the user needs feedback).
   */
  showPlaceholderWhenListening?: boolean;
  /** Optional class on the outer wrapper (in addition to `stack`). */
  className?: string;
};

/**
 * Streaming partial transcript card (same content as the main “Live” section body).
 * Compose under {@link AudioMeter} on any view that should show draft STT text.
 */
export function LivePartialCard(props: Props) {
  const { state, showPlaceholderWhenListening = false, className } = props;
  const voiceListening = state.voiceListening === true;
  const partialRaw =
    typeof state.latestPartial === "string" ? state.latestPartial : "";
  const partial = partialRaw.trim().length > 0 ? partialRaw : null;

  if (!voiceListening) {
    return null;
  }
  if (partial === null && !showPlaceholderWhenListening) {
    return null;
  }

  const showText = partial !== null;
  const wrapClass = ["stack", className].filter(Boolean).join(" ");

  return (
    <div className={wrapClass}>
      <div className="card live">
        <div className="meta">
          <span
            className="badge"
            title="Streaming speech-to-text — not final until you finish the utterance"
          >
            Live
          </span>
          <span title="Draft before the provider commits this segment">
            Draft
          </span>
        </div>
        {showText ? (
          <div className="text">{partial}</div>
        ) : (
          <div className="text live-partial-placeholder muted-transcript">
            Speak — your words appear here as you talk.
          </div>
        )}
        <div className="typing" aria-hidden="true">
          <span className="dot" />
          <span className="dot" />
          <span className="dot" />
        </div>
      </div>
    </div>
  );
}
