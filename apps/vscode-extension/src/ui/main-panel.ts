import * as vscode from "vscode";

import type { MainPanelStore } from "./main-panel-store";

/** VS Code contributed view id (package.json); stable for user layouts and commands. */
const mainPanelViewId = "vocode.transcriptPanel";

type PanelConfigMessage = {
  type: "panelConfig";
  voiceVadDebug: boolean;
  voiceSidecarLogProtocol: boolean;
};

function readPanelConfigMessage(): PanelConfigMessage {
  const c = vscode.workspace.getConfiguration("vocode");
  return {
    type: "panelConfig",
    voiceVadDebug: c.get<boolean>("voiceVadDebug") === true,
    voiceSidecarLogProtocol: c.get<boolean>("voiceSidecarLogProtocol") === true,
  };
}

/**
 * Webview provider for the extension’s main sidebar panel (voice transcript UI).
 * Loads the React bundle from `dist/webview/main-panel.{js,css}`.
 */
export class MainPanelViewProvider
  implements vscode.WebviewViewProvider, vscode.Disposable
{
  private view?: vscode.WebviewView;
  private readonly unsubscribe: () => void;

  constructor(
    private readonly extensionUri: vscode.Uri,
    private readonly store: MainPanelStore,
  ) {
    this.unsubscribe = this.store.onDidChange(() => {
      this.postState();
    });
  }

  resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken,
  ): void {
    this.view = webviewView;
    const webviewRoot = vscode.Uri.joinPath(
      this.extensionUri,
      "dist",
      "webview",
    );
    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this.extensionUri],
    };
    webviewView.webview.html = this.getHtml(webviewView.webview, webviewRoot);
    webviewView.onDidChangeVisibility(() => {
      if (webviewView.visible) {
        this.postState();
      }
    });
    this.postState();

    const wv = webviewView.webview;
    const disposables: vscode.Disposable[] = [];

    disposables.push(
      wv.onDidReceiveMessage((msg: unknown) => {
        if (!msg || typeof msg !== "object") {
          return;
        }
        const m = msg as Record<string, unknown>;
        if (m.type === "requestPanelConfig") {
          void wv.postMessage(readPanelConfigMessage());
          return;
        }
        if (m.type === "setPanelConfig") {
          const patch = m.patch;
          if (!patch || typeof patch !== "object") {
            return;
          }
          const p = patch as Record<string, unknown>;
          const target = vscode.workspace.workspaceFolders?.length
            ? vscode.ConfigurationTarget.Workspace
            : vscode.ConfigurationTarget.Global;
          const config = vscode.workspace.getConfiguration("vocode");
          void (async () => {
            try {
              if (
                "voiceVadDebug" in p &&
                typeof p.voiceVadDebug === "boolean"
              ) {
                await config.update("voiceVadDebug", p.voiceVadDebug, target);
              }
              if (
                "voiceSidecarLogProtocol" in p &&
                typeof p.voiceSidecarLogProtocol === "boolean"
              ) {
                await config.update(
                  "voiceSidecarLogProtocol",
                  p.voiceSidecarLogProtocol,
                  target,
                );
              }
            } finally {
              void wv.postMessage(readPanelConfigMessage());
            }
          })();
          return;
        }
        if (m.type === "openExtensionSettings") {
          void vscode.commands.executeCommand(
            "workbench.action.openSettings",
            "vocode",
          );
        }
      }),
    );

    disposables.push(
      vscode.workspace.onDidChangeConfiguration((e) => {
        if (e.affectsConfiguration("vocode")) {
          void wv.postMessage(readPanelConfigMessage());
        }
      }),
    );

    webviewView.onDidDispose(() => {
      for (const d of disposables) {
        d.dispose();
      }
    });
  }

  private postState(): void {
    if (!this.view) {
      return;
    }
    const snapshot = this.store.getSnapshot();
    const plain = JSON.parse(JSON.stringify(snapshot)) as Record<
      string,
      unknown
    >;
    void this.view.webview.postMessage({
      type: "update",
      state: plain,
    });
  }

  private getHtml(webview: vscode.Webview, webviewRoot: vscode.Uri): string {
    const scriptUri = webview.asWebviewUri(
      vscode.Uri.joinPath(webviewRoot, "main-panel.js"),
    );
    const styleUri = webview.asWebviewUri(
      vscode.Uri.joinPath(webviewRoot, "main-panel.css"),
    );
    const csp = [
      "default-src 'none';",
      `style-src ${webview.cspSource} 'unsafe-inline';`,
      `script-src ${webview.cspSource};`,
    ].join(" ");

    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta http-equiv="Content-Security-Policy" content="${csp}" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link href="${styleUri}" rel="stylesheet" />
</head>
<body>
  <div id="root"></div>
  <script type="module" src="${scriptUri}"></script>
</body>
</html>`;
  }

  dispose(): void {
    this.unsubscribe();
  }
}

/** Pass to `registerWebviewViewProvider` — must match `package.json` `views` id. */
export const mainPanelViewType = mainPanelViewId;
