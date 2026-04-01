import { useLayoutEffect, useRef } from "react";

import type { PanelState } from "../types";
import { fmtTime } from "../util";

export function ChatSection({ state }: { state: PanelState }) {
  const items = Array.isArray(state.qaHistory) ? state.qaHistory : [];
  /** Store is newest-first; render oldest → newest so the latest sits at the bottom. */
  const chronological = [...items].reverse();
  const scrollRef = useRef<HTMLDivElement>(null);
  const qaScrollDigest = items
    .map((q) => `${q.receivedAt}\0${q.question}\0${q.answerText}`)
    .join("\n");

  useLayoutEffect(() => {
    const el = scrollRef.current;
    if (!el || qaScrollDigest.length === 0) {
      return;
    }
    el.scrollTop = el.scrollHeight;
  }, [qaScrollDigest]);

  return (
    <section className="panel-section chat-section">
      <h1>Chat</h1>
      {chronological.length > 0 ? (
        <div
          ref={scrollRef}
          className="chat-thread"
          role="log"
          aria-live="polite"
          aria-relevant="additions"
        >
          {chronological.map((qa) => (
            <div
              key={`qa-${qa.receivedAt}-${qa.question}`}
              className="chat-pair"
            >
              <div className="chat-row chat-row-user">
                <div className="chat-bubble chat-bubble-user">
                  <div className="chat-bubble-text">{qa.question}</div>
                  <div className="chat-bubble-meta">
                    {fmtTime(qa.receivedAt)}
                  </div>
                </div>
              </div>
              <div className="chat-row chat-row-agent">
                <div className="chat-bubble chat-bubble-agent">
                  <div className="chat-bubble-text">
                    {qa.answerText.trim().length > 0 ? qa.answerText : "…"}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : null}
    </section>
  );
}
