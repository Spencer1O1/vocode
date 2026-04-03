import type { EditLocationMap } from "../directives/navigation/execute-navigation-intent";

/** Live host command UI for the voice sidebar row that owns the in-flight transcript RPC. */
export type CommandApplyUiHandlers = {
  readonly pendingId: number;
  readonly onStart: (commandLine: string) => void;
  readonly onOutput: (chunk: string) => void;
};

/** Shared mutable state while applying one transcript result (edits + navigation). */
export type TranscriptApplyContext = {
  activeDocumentPath: string;
  editLocations: EditLocationMap;
  commandApplyUi?: CommandApplyUiHandlers;
};
