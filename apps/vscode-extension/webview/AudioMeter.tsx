import { useLayoutEffect, useRef } from "react";

import type { PanelState } from "./types";

function drawWaveform(
  canvas: HTMLCanvasElement | null,
  samples: readonly number[],
) {
  if (!canvas?.getContext) {
    return;
  }
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    return;
  }
  const w = canvas.width;
  const h = canvas.height;
  ctx.clearRect(0, 0, w, h);
  const fg =
    getComputedStyle(document.body)
      .getPropertyValue("--vscode-textLink-foreground")
      .trim() || "#3794ff";
  const arr = Array.isArray(samples) ? samples : [];
  if (arr.length === 0) {
    ctx.strokeStyle = fg;
    ctx.globalAlpha = 0.22;
    ctx.beginPath();
    ctx.moveTo(0, h - 2);
    ctx.lineTo(w, h - 2);
    ctx.stroke();
    ctx.globalAlpha = 1;
    return;
  }
  const n = arr.length;
  const gap = 1;
  const barW = Math.max(1, (w - (n - 1) * gap) / n);
  let x = 0;
  for (let i = 0; i < n; i++) {
    const v = typeof arr[i] === "number" ? arr[i] : 0;
    const bh = Math.max(1, Math.min(1, v) * (h - 4));
    ctx.fillStyle = fg;
    ctx.globalAlpha = 0.75 + 0.25 * Math.min(1, v);
    ctx.fillRect(x, h - bh, barW, bh);
    x += barW + gap;
  }
  ctx.globalAlpha = 1;
}

export function AudioMeter(props: { state: PanelState }) {
  const { state } = props;
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const voiceListening = state.voiceListening === true;
  const am = state.audioMeter;
  const rms = typeof am.rms === "number" ? am.rms : 0;
  const speaking = am.speaking === true;
  const pct = Math.round(Math.min(1, Math.max(0, rms)) * 100);

  useLayoutEffect(() => {
    drawWaveform(canvasRef.current, am.waveform ?? []);
  }, [am.waveform]);

  return (
    <div className="meter card">
      <div className="meta">
        <span className="badge">
          {!voiceListening ? "Idle" : speaking ? "Speaking" : "Quiet"}
        </span>
        <span>{!voiceListening ? "Not listening" : "Input level"}</span>
      </div>
      <div className="meter-bar">
        <div className="meter-fill" style={{ width: `${pct}%` }} />
      </div>
      <canvas
        ref={canvasRef}
        className="wave-canvas"
        width={320}
        height={44}
        aria-label="Recent level"
      />
    </div>
  );
}
