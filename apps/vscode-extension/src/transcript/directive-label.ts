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
    a.kind === "replace_range" ||
    a.kind === "replace_file" ||
    a.kind === "create_file" ||
    a.kind === "append_to_file"
  ) {
    return a.path;
  }
  return undefined;
}

function editSummary(d: VoiceTranscriptDirective): string {
  const ed = d.editDirective;
  if (!ed || ed.kind !== "success" || !ed.actions?.length) {
    return "Edit";
  }
  const a = ed.actions[0] as EditAction;
  switch (a.kind) {
    case "replace_range":
      return "Scoped edit";
    case "replace_file":
      return "Replace file";
    case "replace_between_anchors":
      return "Edit (anchors)";
    case "create_file":
      return "Create file";
    case "append_to_file":
      return "Append to file";
    default:
      return "Edit";
  }
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
      const base = editSummary(d);
      return p !== undefined
        ? `${n}. ${base}: ${path.basename(p)}`
        : `${n}. ${base}`;
    }
    case "navigate":
      return `${n}. ${navigationSummary(d.navigationDirective)}`;
    case "undo": {
      const scope = d.undoDirective?.scope ?? "undo";
      return `${n}. Undo (${scope})`;
    }
    case "rename": {
      const p = d.renameDirective?.path
        ? path.basename(d.renameDirective.path)
        : "file";
      const newName = d.renameDirective?.newName ?? "rename";
      return `${n}. Rename in ${p} → ${newName}`;
    }
    case "format": {
      const p = d.formatDirective?.path
        ? path.basename(d.formatDirective.path)
        : "file";
      const scope = d.formatDirective?.scope ?? "document";
      return `${n}. Format (${scope}): ${p}`;
    }
    case "code_action": {
      const p = d.codeActionDirective?.path
        ? path.basename(d.codeActionDirective.path)
        : "file";
      const kind = d.codeActionDirective?.actionKind ?? "code action";
      return `${n}. ${kind}: ${p}`;
    }
    default:
      return `${n}. Directive`;
  }
}
