interface UndoDeps {
  hasActiveEditor: () => boolean;
  executeUndo: () => PromiseLike<void>;
  showWarning: (message: string) => void;
}

export async function runUndoLastEditWithDeps(deps: UndoDeps): Promise<void> {
  if (!deps.hasActiveEditor()) {
    deps.showWarning(
      "Open a text editor before running Vocode: Undo Last Edit.",
    );
    return;
  }

  await deps.executeUndo();
}
