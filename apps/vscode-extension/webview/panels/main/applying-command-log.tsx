import { useEffect, useRef } from "react";

import type { PendingRow } from "../../types";

/**
 * Live shell invocation + streamed stdout/stderr while the host applies a command directive.
 */
export function ApplyingCommandLog({ row }: { row: PendingRow }) {
  const preRef = useRef<HTMLPreElement>(null);
  const line = row.applyingCommandLine?.trim();
  const out = row.applyingCommandOutput;

  useEffect(() => {
    const el = preRef.current;
    if (!el || out === undefined || out === "") {
      return;
    }
    el.scrollTop = el.scrollHeight;
  }, [out]);

  if (!line && (out === undefined || out === "")) {
    return null;
  }
  return (
    <div
      className="applying-command-log-block"
      role="region"
      aria-label="Command output"
    >
      {line ? (
        <div className="applying-command-log-line" title={line}>
          <span className="applying-command-log-label">Running</span>
          <code className="applying-command-log-cmd">{line}</code>
        </div>
      ) : null}
      {out !== undefined && out !== "" ? (
        <pre ref={preRef} className="applying-command-log-pre">
          {out}
        </pre>
      ) : null}
    </div>
  );
}
