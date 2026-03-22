// AUTO-GENERATED. DO NOT EDIT.


export interface ReplaceBetweenAnchorsAction {
  kind: "replace_between_anchors";
  path: string;
  anchor: {
    before: string;
    after: string;
  };
  newText: string;
}

export type EditAction = ReplaceBetweenAnchorsAction;

export interface PingParams {}

export interface PingResult {
  message: "pong";
}

export interface EditApplyParams {
  instruction: string;
}

export interface EditApplyResult {
  actions: EditAction[];
}