import path from "node:path";
import type {
  EditAction,
  NavigationAction,
  NavigationDirective,
  VoiceTranscriptDirective,
} from "@vocode/protocol";

function firstEditPath(
  editDirective: VoiceTranscriptDirective["editDirective"],
): string | undefined {
  if (
    !editDirective ||
    editDirective.kind !== "success" ||
    !editDirective.actions?.length
  ) {
    return undefined;
  }
  const a = editDirective.actions[0] as EditAction;
  if (
    a.kind === "replace_between_anchors" ||
    a.kind === "create_file" ||
    a.kind === "append_to_file"
  ) {
    return a.path;
  }
  return undefined;
}

function navigationSummary(nav: NavigationDirective | undefined): string {
  if (!nav) {
    return "Navigate";
  }
  if (nav.kind === "noop") {
    return "Navigate (noop)";
  }
  const act: NavigationAction = nav.action;
  switch (act.kind) {
    case "open_file":
      return `Open ${path.basename(act.openFile.path)}`;
    case "reveal_symbol":
      return `Reveal ${act.revealSymbol.symbolName}`;
    case "move_cursor":
      return "Move cursor";
    case "select_range":
      return "Select range";
    case "reveal_edit":
      return `Reveal edit ${act.revealEdit.editId}`;
    default:
      return "Navigate";
  }
}

/**
 * Short host-facing label for sidebar checklist (one line per directive).
 */
export function directiveApplyLabel(
  d: VoiceTranscriptDirective,
  index: number,
): string {
  const n = index + 1;
  switch (d.kind) {
    case "command": {
      const cmd = d.commandDirective?.command?.trim() || "command";
      const args = d.commandDirective?.args?.filter(Boolean).join(" ") ?? "";
      return args.length > 0 ? `${n}. ${cmd} ${args}` : `${n}. ${cmd}`;
    }
    case "edit": {
      const p = firstEditPath(d.editDirective);
      return p !== undefined ? `${n}. Edit ${path.basename(p)}` : `${n}. Edit`;
    }
    case "navigate":
      return `${n}. ${navigationSummary(d.navigationDirective)}`;
    case "undo": {
      const scope = d.undoDirective?.scope ?? "undo";
      return `${n}. Undo (${scope})`;
    }
    default:
      return `${n}. Directive`;
  }
}
